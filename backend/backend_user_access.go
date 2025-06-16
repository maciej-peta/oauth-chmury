package main

import (
	"database/sql"
)

// used multiple smaller functions in here first but then I realized I could simplify
// it to just one request, which would shorten meantime to complete conversion way down
func canUserConvert(db *sql.DB, authID string) (bool, int, error) { //todo: clean up this method so it doesnt use bool
	var maxInteractions, maxFileSizeMB, recentConversions int

	query := `
        SELECT 
            at.max_interactions,
            at.max_file_size_mb,
            COALESCE(c.count, 0) as recent_conversions
        FROM users u
        JOIN account_types at ON u.account_type_id = at.account_type_id
        LEFT JOIN (
            SELECT user_id, COUNT(*) as count
            FROM conversions
            WHERE created_at >= NOW() - INTERVAL '24 hours'
            GROUP BY user_id
        ) c ON u.user_id = c.user_id
        WHERE u.auth_id = $1
    `

	err := db.QueryRow(query, authID).Scan(&maxInteractions, &maxFileSizeMB, &recentConversions)
	if err != nil {
		return false, 0, err
	}

	if recentConversions >= maxInteractions {
		return false, maxFileSizeMB, nil
	}

	return true, maxFileSizeMB, nil
}

func insertConversion(
	db *sql.DB,
	authID string,
	startingType string,
	resultingType string,
	fileSizeKB int,
) error {
	query := `
        INSERT INTO conversions (user_id, starting_type, resulting_type, file_size_kb)
        VALUES (
            (SELECT user_id FROM users WHERE auth_id = $1),
            $2, $3, $4
        )
    `
	_, err := db.Exec(query, authID, startingType, resultingType, fileSizeKB)
	return err
}
