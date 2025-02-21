//go:build linux && cgo && !agent

package cluster

// The code below was generated by lxd-generate - DO NOT EDIT!

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/shared/api"
)

var _ = api.ServerEnvironment{}

var warningObjects = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  ORDER BY warnings.uuid
`)

var warningObjectsByUUID = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( warnings.uuid = ? )
  ORDER BY warnings.uuid
`)

var warningObjectsByProject = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( coalesce(project, '') = ? )
  ORDER BY warnings.uuid
`)

var warningObjectsByStatus = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( warnings.status = ? )
  ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCode = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( coalesce(node, '') = ? AND warnings.type_code = ? )
  ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCodeAndProject = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( coalesce(node, '') = ? AND warnings.type_code = ? AND coalesce(project, '') = ? )
  ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings
  LEFT JOIN nodes ON warnings.node_id = nodes.id
  LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE ( coalesce(node, '') = ? AND warnings.type_code = ? AND coalesce(project, '') = ? AND coalesce(warnings.entity_type_code, -1) = ? AND coalesce(warnings.entity_id, -1) = ? )
  ORDER BY warnings.uuid
`)

var warningDeleteByUUID = RegisterStmt(`
DELETE FROM warnings WHERE uuid = ?
`)

var warningDeleteByEntityTypeCodeAndEntityID = RegisterStmt(`
DELETE FROM warnings WHERE entity_type_code = ? AND entity_id = ?
`)

var warningID = RegisterStmt(`
SELECT warnings.id FROM warnings
  WHERE warnings.uuid = ?
`)

// warningColumns returns a string of column names to be used with a SELECT statement for the entity.
// Use this function when building statements to retrieve database entries matching the Warning entity.
func warningColumns() string {
	return "warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count"
}

// getWarnings can be used to run handwritten sql.Stmts to return a slice of objects.
func getWarnings(ctx context.Context, stmt *sql.Stmt, args ...any) ([]Warning, error) {
	objects := make([]Warning, 0)

	dest := func(scan func(dest ...any) error) error {
		w := Warning{}
		err := scan(&w.ID, &w.Node, &w.Project, &w.EntityTypeCode, &w.EntityID, &w.UUID, &w.TypeCode, &w.Status, &w.FirstSeenDate, &w.LastSeenDate, &w.UpdatedDate, &w.LastMessage, &w.Count)
		if err != nil {
			return err
		}

		objects = append(objects, w)

		return nil
	}

	err := query.SelectObjects(ctx, stmt, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	return objects, nil
}

// getWarnings can be used to run handwritten query strings to return a slice of objects.
func getWarningsRaw(ctx context.Context, tx *sql.Tx, sql string, args ...any) ([]Warning, error) {
	objects := make([]Warning, 0)

	dest := func(scan func(dest ...any) error) error {
		w := Warning{}
		err := scan(&w.ID, &w.Node, &w.Project, &w.EntityTypeCode, &w.EntityID, &w.UUID, &w.TypeCode, &w.Status, &w.FirstSeenDate, &w.LastSeenDate, &w.UpdatedDate, &w.LastMessage, &w.Count)
		if err != nil {
			return err
		}

		objects = append(objects, w)

		return nil
	}

	err := query.Scan(ctx, tx, sql, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	return objects, nil
}

// GetWarnings returns all available warnings.
// generator: warning GetMany
func GetWarnings(ctx context.Context, tx *sql.Tx, filters ...WarningFilter) ([]Warning, error) {
	var err error

	// Result slice.
	objects := make([]Warning, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	var sqlStmt *sql.Stmt
	args := []any{}
	queryParts := [2]string{}

	if len(filters) == 0 {
		sqlStmt, err = Stmt(tx, warningObjects)
		if err != nil {
			return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
		}
	}

	for i, filter := range filters {
		if filter.Node != nil && filter.TypeCode != nil && filter.Project != nil && filter.EntityTypeCode != nil && filter.EntityID != nil && filter.ID == nil && filter.UUID == nil && filter.Status == nil {
			args = append(args, []any{filter.Node, filter.TypeCode, filter.Project, filter.EntityTypeCode, filter.EntityID}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.Node != nil && filter.TypeCode != nil && filter.Project != nil && filter.ID == nil && filter.UUID == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
			args = append(args, []any{filter.Node, filter.TypeCode, filter.Project}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByNodeAndTypeCodeAndProject)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByNodeAndTypeCodeAndProject\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByNodeAndTypeCodeAndProject)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.Node != nil && filter.TypeCode != nil && filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
			args = append(args, []any{filter.Node, filter.TypeCode}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByNodeAndTypeCode)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByNodeAndTypeCode\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByNodeAndTypeCode)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.UUID != nil && filter.ID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
			args = append(args, []any{filter.UUID}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByUUID)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByUUID\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByUUID)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.Status != nil && filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil {
			args = append(args, []any{filter.Status}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByStatus)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByStatus\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByStatus)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.Project != nil && filter.ID == nil && filter.UUID == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
			args = append(args, []any{filter.Project}...)
			if len(filters) == 1 {
				sqlStmt, err = Stmt(tx, warningObjectsByProject)
				if err != nil {
					return nil, fmt.Errorf("Failed to get \"warningObjectsByProject\" prepared statement: %w", err)
				}

				break
			}

			query, err := StmtString(warningObjectsByProject)
			if err != nil {
				return nil, fmt.Errorf("Failed to get \"warningObjects\" prepared statement: %w", err)
			}

			parts := strings.SplitN(query, "ORDER BY", 2)
			if i == 0 {
				copy(queryParts[:], parts)
				continue
			}

			_, where, _ := strings.Cut(parts[0], "WHERE")
			queryParts[0] += "OR" + where
		} else if filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
			return nil, fmt.Errorf("Cannot filter on empty WarningFilter")
		} else {
			return nil, fmt.Errorf("No statement exists for the given Filter")
		}
	}

	// Select.
	if sqlStmt != nil {
		objects, err = getWarnings(ctx, sqlStmt, args...)
	} else {
		queryStr := strings.Join(queryParts[:], "ORDER BY")
		objects, err = getWarningsRaw(ctx, tx, queryStr, args...)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	return objects, nil
}

// GetWarning returns the warning with the given key.
// generator: warning GetOne-by-UUID
func GetWarning(ctx context.Context, tx *sql.Tx, uuid string) (*Warning, error) {
	filter := WarningFilter{}
	filter.UUID = &uuid

	objects, err := GetWarnings(ctx, tx, filter)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	switch len(objects) {
	case 0:
		return nil, api.StatusErrorf(http.StatusNotFound, "Warning not found")
	case 1:
		return &objects[0], nil
	default:
		return nil, fmt.Errorf("More than one \"warnings\" entry matches")
	}
}

// DeleteWarning deletes the warning matching the given key parameters.
// generator: warning DeleteOne-by-UUID
func DeleteWarning(ctx context.Context, tx *sql.Tx, uuid string) error {
	stmt, err := Stmt(tx, warningDeleteByUUID)
	if err != nil {
		return fmt.Errorf("Failed to get \"warningDeleteByUUID\" prepared statement: %w", err)
	}

	result, err := stmt.Exec(uuid)
	if err != nil {
		return fmt.Errorf("Delete \"warnings\": %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "Warning not found")
	} else if n > 1 {
		return fmt.Errorf("Query deleted %d Warning rows instead of 1", n)
	}

	return nil
}

// DeleteWarnings deletes the warning matching the given key parameters.
// generator: warning DeleteMany-by-EntityTypeCode-and-EntityID
func DeleteWarnings(ctx context.Context, tx *sql.Tx, entityTypeCode int, entityID int) error {
	stmt, err := Stmt(tx, warningDeleteByEntityTypeCodeAndEntityID)
	if err != nil {
		return fmt.Errorf("Failed to get \"warningDeleteByEntityTypeCodeAndEntityID\" prepared statement: %w", err)
	}

	result, err := stmt.Exec(entityTypeCode, entityID)
	if err != nil {
		return fmt.Errorf("Delete \"warnings\": %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	return nil
}

// GetWarningID return the ID of the warning with the given key.
// generator: warning ID
func GetWarningID(ctx context.Context, tx *sql.Tx, uuid string) (int64, error) {
	stmt, err := Stmt(tx, warningID)
	if err != nil {
		return -1, fmt.Errorf("Failed to get \"warningID\" prepared statement: %w", err)
	}

	row := stmt.QueryRowContext(ctx, uuid)
	var id int64
	err = row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, api.StatusErrorf(http.StatusNotFound, "Warning not found")
	}

	if err != nil {
		return -1, fmt.Errorf("Failed to get \"warnings\" ID: %w", err)
	}

	return id, nil
}

// WarningExists checks if a warning with the given key exists.
// generator: warning Exists
func WarningExists(ctx context.Context, tx *sql.Tx, uuid string) (bool, error) {
	_, err := GetWarningID(ctx, tx, uuid)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
