#!/bin/bash

# Create test database
PGPASSWORD=api_password psql -h localhost -U api_user -d postgres -c "DROP DATABASE IF EXISTS digital_identity_test;"
PGPASSWORD=api_password psql -h localhost -U api_user -d postgres -c "CREATE DATABASE digital_identity_test;"

# Grant privileges
PGPASSWORD=api_password psql -h localhost -U api_user -d digital_identity_test -c "GRANT ALL PRIVILEGES ON DATABASE digital_identity_test TO api_user;"