package models

import (
	"context"
	"time"
)

// createTables create tables in the database
func createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone_number VARCHAR(18) UNIQUE NOT NULL,
			device_token Text NOT NULL,
			pin VARCHAR(100) NOT NULL,
			quota BIGINT DEFAULT 0 NOT NULL,
			locked BOOLEAN DEFAULT FALSE NOT NULL,
			premium BOOLEAN DEFAULT FALSE NOT NULL,
			photo VARCHAR(200),
			is_active BOOLEAN DEFAULT TRUE NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS users_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			device_token Text NOT NULL,
			phone_number VARCHAR(18) NOT NULL,
			activity VARCHAR(50) NOT NULL, -- 'login', 'logout', etc.	
			metadata JSONB, -- optional extra info (e.g. payment method, ref number)
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_log_user FOREIGN KEY (user_id)
				REFERENCES users (id)
				ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_lookup ON users (phone_number, is_active, quota, locked, premium);`,
		`CREATE INDEX IF NOT EXISTS idx_users_logs_created_at ON users_logs(created_at, user_id, phone_number);`,
	}
	for _, query := range queries {
		if _, err := DB.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
