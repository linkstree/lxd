package device

import (
	"fmt"
	"net"
	"strings"

	deviceConfig "github.com/lxc/lxd/lxd/device/config"
	"github.com/lxc/lxd/lxd/instance"
	"github.com/lxc/lxd/lxd/instance/instancetype"
	"github.com/lxc/lxd/lxd/ip"
	"github.com/lxc/lxd/lxd/network"
	"github.com/lxc/lxd/lxd/util"
	"github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/validate"
)

const ipvlanModeL3S = "l3s"
const ipvlanModeL2 = "l2"

type nicIPVLAN struct {
	deviceCommon
}

// CanHotPlug returns whether the device can be managed whilst the instance is running,.
func (d *nicIPVLAN) CanHotPlug() bool {
	return false
}

// validateConfig checks the supplied config for correctness.
func (d *nicIPVLAN) validateConfig(instConf instance.ConfigReader) error {
	if !instanceSupported(instConf.Type(), instancetype.Container) {
		return ErrUnsupportedDevType
	}

	requiredFields := []string{"parent"}
	optionalFields := []string{
		"name",
		"mtu",
		"hwaddr",
		"vlan",
		"ipv4.gateway",
		"ipv6.gateway",
		"ipv4.host_table",
		"ipv6.host_table",
		"gvrp",
	}

	rules := nicValidationRules(requiredFields, optionalFields, instConf)
	rules["gvrp"] = validate.Optional(validate.IsBool)
	rules["ipv4.address"] = func(value string) error {
		if value == "" {
			return nil
		}

		if d.config["mode"] == ipvlanModeL2 {
			for _, v := range strings.Split(value, ",") {
				v = strings.TrimSpace(v)

				// If valid non-CIDR address specified, append a /24 subnet.
				if validate.IsNetworkAddressV4(v) == nil {
					v = fmt.Sprintf("%s/24", v)
				}

				ip, _, err := net.ParseCIDR(v)
				if err != nil {
					return err
				}

				if ip.To4() == nil {
					return fmt.Errorf("Not an IPv4 CIDR address: %s", v)
				}
			}

			return nil
		}

		return validate.IsListOf(validate.IsNetworkAddressV4)(value)
	}

	rules["ipv6.address"] = func(value string) error {
		if value == "" {
			return nil
		}

		if d.config["mode"] == ipvlanModeL2 {
			for _, v := range strings.Split(value, ",") {
				v = strings.TrimSpace(v)

				// If valid non-CIDR address specified, append a /64 subnet.
				if validate.IsNetworkAddressV6(v) == nil {
					v = fmt.Sprintf("%s/64", v)
				}

				ip, _, err := net.ParseCIDR(v)
				if err != nil {
					return err
				}

				if ip == nil || ip.To4() != nil {
					return fmt.Errorf("Not an IPv6 CIDR address: %s", v)
				}
			}

			return nil
		}

		return validate.IsListOf(validate.IsNetworkAddressV6)(value)
	}

	rules["mode"] = func(value string) error {
		if value == "" {
			return nil
		}

		validModes := []string{ipvlanModeL3S, ipvlanModeL2}
		if !shared.StringInSlice(value, validModes) {
			return fmt.Errorf("Must be one of: %v", strings.Join(validModes, ", "))
		}

		return nil
	}

	if d.config["mode"] == ipvlanModeL2 {
		rules["ipv4.gateway"] = validate.Optional(validate.IsNetworkAddressV4)
		rules["ipv6.gateway"] = validate.Optional(validate.IsNetworkAddressV6)
	}

	err := d.config.Validate(rules)
	if err != nil {
		return err
	}

	if d.config["mode"] == ipvlanModeL2 && d.config["host_table"] != "" {
		return fmt.Errorf("host_table option cannot be used in l2 mode")
	}

	return nil
}

// validateEnvironment checks the runtime environment for correctness.
func (d *nicIPVLAN) validateEnvironment() error {
	if d.inst.Type() == instancetype.Container && d.config["name"] == "" {
		return fmt.Errorf("Requires name property to start")
	}

	extensions := d.state.OS.LXCFeatures
	if !extensions["network_ipvlan"] || !extensions["network_l2proxy"] || !extensions["network_gateway_device_route"] {
		return fmt.Errorf("Requires liblxc has following API extensions: network_ipvlan, network_l2proxy, network_gateway_device_route")
	}

	if !network.InterfaceExists(d.config["parent"]) {
		return fmt.Errorf("Parent device '%s' doesn't exist", d.config["parent"])
	}

	if d.config["parent"] == "" && d.config["vlan"] != "" {
		return fmt.Errorf("The vlan setting can only be used when combined with a parent interface")
	}

	// Only check sysctls for l2proxy if mode is l3s.
	if d.mode() != ipvlanModeL3S {
		return nil
	}

	// Generate effective parent name, including the VLAN part if option used.
	effectiveParentName := network.GetHostDevice(d.config["parent"], d.config["vlan"])

	// If the effective parent doesn't exist and the vlan option is specified, it means we are going to create
	// the VLAN parent at start, and we will configure the needed sysctls so don't need to check them yet.
	if d.config["vlan"] != "" && !network.InterfaceExists(effectiveParentName) {
		return nil
	}

	if d.config["ipv4.address"] != "" {
		// Check necessary sysctls are configured for use with l2proxy parent in IPVLAN l3s mode.
		ipv4FwdPath := fmt.Sprintf("net/ipv4/conf/%s/forwarding", effectiveParentName)
		sysctlVal, err := util.SysctlGet(ipv4FwdPath)
		if err != nil {
			return fmt.Errorf("Error reading net sysctl %s: %w", ipv4FwdPath, err)
		}

		if sysctlVal != "1\n" {
			// Replace . in parent name with / for sysctl formatting.
			return fmt.Errorf("IPVLAN in L3S mode requires sysctl net.ipv4.conf.%s.forwarding=1", strings.Replace(effectiveParentName, ".", "/", -1))
		}
	}

	if d.config["ipv6.address"] != "" {
		// Check necessary sysctls are configured for use with l2proxy parent in IPVLAN l3s mode.
		ipv6FwdPath := fmt.Sprintf("net/ipv6/conf/%s/forwarding", effectiveParentName)
		sysctlVal, err := util.SysctlGet(ipv6FwdPath)
		if err != nil {
			return fmt.Errorf("Error reading net sysctl %s: %w", ipv6FwdPath, err)
		}

		if sysctlVal != "1\n" {
			// Replace . in parent name with / for sysctl formatting.
			return fmt.Errorf("IPVLAN in L3S mode requires sysctl net.ipv6.conf.%s.forwarding=1", strings.Replace(effectiveParentName, ".", "/", -1))
		}

		ipv6ProxyNdpPath := fmt.Sprintf("net/ipv6/conf/%s/proxy_ndp", effectiveParentName)
		sysctlVal, err = util.SysctlGet(ipv6ProxyNdpPath)
		if err != nil {
			return fmt.Errorf("Error reading net sysctl %s: %w", ipv6ProxyNdpPath, err)
		}

		if sysctlVal != "1\n" {
			// Replace . in parent name with / for sysctl formatting.
			return fmt.Errorf("IPVLAN in L3S mode requires sysctl net.ipv6.conf.%s.proxy_ndp=1", strings.Replace(effectiveParentName, ".", "/", -1))
		}
	}

	return nil
}

// Start is run when the instance is starting up (IPVLAN doesn't support hot plugging).
func (d *nicIPVLAN) Start() (*deviceConfig.RunConfig, error) {
	err := d.validateEnvironment()
	if err != nil {
		return nil, err
	}

	// Lock to avoid issues with containers starting in parallel.
	networkCreateSharedDeviceLock.Lock()
	defer networkCreateSharedDeviceLock.Unlock()

	saveData := make(map[string]string)

	// Record a random host name to use to detach the ipvlan interface back onto the host at stop time so we
	// can remove it and not have to rely on the kernel to do it when the namespace is destroyed, as this is
	// not always reliable.
	saveData["host_name"] = network.RandomDevName("lxd")

	// Decide which parent we should use based on VLAN setting.
	parentName := network.GetHostDevice(d.config["parent"], d.config["vlan"])

	statusDev, err := networkCreateVlanDeviceIfNeeded(d.state, d.config["parent"], parentName, d.config["vlan"], shared.IsTrue(d.config["gvrp"]))
	if err != nil {
		return nil, err
	}

	// Record whether we created this device or not so it can be removed on stop.
	saveData["last_state.created"] = fmt.Sprintf("%t", statusDev != "existing")

	mode := d.mode()

	// If we created a VLAN interface, we need to setup the sysctls on that interface for l3s mode l2proxy.
	if statusDev == "created" && mode == ipvlanModeL3S {
		err := d.setupParentSysctls(parentName)
		if err != nil {
			return nil, err
		}
	}

	err = d.volatileSet(saveData)
	if err != nil {
		return nil, err
	}

	runConf := deviceConfig.RunConfig{}
	nic := []deviceConfig.RunConfigItem{
		{Key: "name", Value: d.config["name"]},
		{Key: "type", Value: "ipvlan"},
		{Key: "flags", Value: "up"},
		{Key: "ipvlan.mode", Value: mode},
		{Key: "ipvlan.isolation", Value: "bridge"},
		{Key: "link", Value: parentName},
	}

	// Enable l2proxy for l3s mode.
	if mode == ipvlanModeL3S {
		nic = append(nic, deviceConfig.RunConfigItem{Key: "l2proxy", Value: "1"})
	}

	if d.config["mtu"] != "" {
		nic = append(nic, deviceConfig.RunConfigItem{Key: "mtu", Value: d.config["mtu"]})
	}

	if d.config["ipv4.address"] != "" {
		for _, addr := range strings.Split(d.config["ipv4.address"], ",") {
			addr = strings.TrimSpace(addr)

			if mode == ipvlanModeL3S {
				addr = fmt.Sprintf("%s/32", addr)
			}

			if mode == ipvlanModeL2 && validate.IsNetworkAddressV4(addr) == nil {
				addr = fmt.Sprintf("%s/24", addr)
			}

			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv4.address", Value: addr})
		}

		if mode == ipvlanModeL3S && nicHasAutoGateway(d.config["ipv4.gateway"]) {
			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv4.gateway", Value: "dev"})
		}

		if mode == ipvlanModeL2 && d.config["ipv4.gateway"] != "" {
			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv4.gateway", Value: d.config["ipv4.gateway"]})
		}
	}

	if d.config["ipv6.address"] != "" {
		for _, addr := range strings.Split(d.config["ipv6.address"], ",") {
			addr = strings.TrimSpace(addr)

			if mode == ipvlanModeL3S {
				addr = fmt.Sprintf("%s/128", addr)
			}

			if mode == "l2" && validate.IsNetworkAddressV6(addr) == nil {
				addr = fmt.Sprintf("%s/64", addr)
			}

			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv6.address", Value: addr})
		}

		if mode == ipvlanModeL3S && nicHasAutoGateway(d.config["ipv6.gateway"]) {
			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv6.gateway", Value: "dev"})
		}

		if mode == ipvlanModeL2 && d.config["ipv6.gateway"] != "" {
			nic = append(nic, deviceConfig.RunConfigItem{Key: "ipv6.gateway", Value: d.config["ipv6.gateway"]})
		}
	}

	runConf.NetworkInterface = nic
	runConf.PostHooks = append(runConf.PostHooks, d.postStart)
	return &runConf, nil
}

// setupParentSysctls configures the required sysctls on the parent to allow l2proxy to work.
// Because of our policy not to modify sysctls on existing interfaces, this should only be called
// if we created the parent interface.
func (d *nicIPVLAN) setupParentSysctls(parentName string) error {
	if d.config["ipv4.address"] != "" {
		// Set necessary sysctls for use with l2proxy parent in IPVLAN l3s mode.
		ipv4FwdPath := fmt.Sprintf("net/ipv4/conf/%s/forwarding", parentName)
		err := util.SysctlSet(ipv4FwdPath, "1")
		if err != nil {
			return fmt.Errorf("Error setting net sysctl %s: %w", ipv4FwdPath, err)
		}
	}

	if d.config["ipv6.address"] != "" {
		// Set necessary sysctls use with l2proxy parent in IPVLAN l3s mode.
		ipv6FwdPath := fmt.Sprintf("net/ipv6/conf/%s/forwarding", parentName)
		err := util.SysctlSet(ipv6FwdPath, "1")
		if err != nil {
			return fmt.Errorf("Error setting net sysctl %s: %w", ipv6FwdPath, err)
		}

		ipv6ProxyNdpPath := fmt.Sprintf("net/ipv6/conf/%s/proxy_ndp", parentName)
		err = util.SysctlSet(ipv6ProxyNdpPath, "1")
		if err != nil {
			return fmt.Errorf("Error setting net sysctl %s: %w", ipv6ProxyNdpPath, err)
		}
	}

	return nil
}

// postStart is run after the instance is started.
func (d *nicIPVLAN) postStart() error {
	if d.config["ipv4.address"] != "" {
		// Add static routes to instance IPs to custom routing tables if specified.
		// This is in addition to the static route added by liblxc to the main routing table.
		if d.config["ipv4.host_table"] != "" {
			for _, addr := range strings.Split(d.config["ipv4.address"], ",") {
				addr = strings.TrimSpace(addr)
				r := &ip.Route{
					DevName: "lo",
					Route:   fmt.Sprintf("%s/32", addr),
					Table:   d.config["ipv4.host_table"],
					Family:  ip.FamilyV4,
				}

				err := r.Add()
				if err != nil {
					return err
				}
			}
		}
	}

	if d.config["ipv6.address"] != "" {
		// Add static routes to instance IPs to custom routing tables if specified.
		// This is in addition to the static route added by liblxc to the main routing table.
		if d.config["ipv6.host_table"] != "" {
			for _, addr := range strings.Split(d.config["ipv6.address"], ",") {
				addr = strings.TrimSpace(addr)
				r := &ip.Route{
					DevName: "lo",
					Route:   fmt.Sprintf("%s/128", addr),
					Table:   d.config["ipv6.host_table"],
					Family:  ip.FamilyV6,
				}

				err := r.Add()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Stop is run when the device is removed from the instance.
func (d *nicIPVLAN) Stop() (*deviceConfig.RunConfig, error) {
	v := d.volatileGet()
	runConf := deviceConfig.RunConfig{
		PostHooks: []func() error{d.postStop},
	}

	// Add instruction for removal of ipvlan interface back to host if set.
	if v["host_name"] != "" {
		runConf.NetworkInterface = []deviceConfig.RunConfigItem{
			{Key: "link", Value: v["host_name"]},
		}
	}

	return &runConf, nil
}

// postStop is run after the device is removed from the instance.
func (d *nicIPVLAN) postStop() error {
	defer func() {
		_ = d.volatileSet(map[string]string{
			"last_state.created": "",
			"host_name":          "",
		})
	}()

	v := d.volatileGet()

	networkVethFillFromVolatile(d.config, v)

	errs := []error{}

	// Delete host-side detached interface if not removed by liblxc.
	if network.InterfaceExists(d.config["host_name"]) {
		err := network.InterfaceRemove(d.config["host_name"])
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed to remove interface %q: %w", d.config["host_name"], err))
		}
	}

	if d.config["ipv4.address"] != "" {
		// Remove static routes to instance IPs to custom routing tables if specified.
		if d.config["ipv4.host_table"] != "" {
			for _, addr := range strings.Split(d.config["ipv4.address"], ",") {
				addr = strings.TrimSpace(addr)
				r := &ip.Route{
					DevName: "lo",
					Route:   fmt.Sprintf("%s/32", addr),
					Table:   d.config["ipv4.host_table"],
					Family:  ip.FamilyV4,
				}

				err := r.Delete()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	if d.config["ipv6.address"] != "" {
		// Remove static routes to instance IPs to custom routing tables if specified.
		if d.config["ipv6.host_table"] != "" {
			for _, addr := range strings.Split(d.config["ipv6.address"], ",") {
				addr = strings.TrimSpace(addr)
				r := &ip.Route{
					DevName: "lo",
					Route:   fmt.Sprintf("%s/128", addr),
					Table:   d.config["ipv6.host_table"],
					Family:  ip.FamilyV6,
				}

				err := r.Delete()
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	// This will delete the parent interface if we created it for VLAN parent.
	if shared.IsTrue(v["last_state.created"]) {
		parentName := network.GetHostDevice(d.config["parent"], d.config["vlan"])
		err := networkRemoveInterfaceIfNeeded(d.state, parentName, d.inst, d.config["parent"], d.config["vlan"])
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// mode returns the ipvlan mode to use.
func (d *nicIPVLAN) mode() string {
	if d.config["mode"] == ipvlanModeL2 {
		return ipvlanModeL2
	}

	return ipvlanModeL3S
}
