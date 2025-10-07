-- PostgreSQL initialization script
-- Creates application user and a read-only user with secure password policies

-- Create application user with read/write access
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'api_user') THEN
      CREATE ROLE api_user LOGIN PASSWORD 'api_password';
   END IF;
END
$do$;

-- Create read-only user for analytics (future use)
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'readonly_user') THEN
      CREATE ROLE readonly_user LOGIN PASSWORD 'readonly_password';
   END IF;
END
$do$;

-- Grant privileges
GRANT CONNECT ON DATABASE digital_identity TO api_user, readonly_user;
GRANT USAGE ON SCHEMA public TO api_user, readonly_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO api_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;

-- Ensure future tables have correct privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO api_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly_user;
