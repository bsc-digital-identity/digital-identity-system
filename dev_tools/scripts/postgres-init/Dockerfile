# This Dockerfile is optional for PostgreSQL, as the official image supports init scripts via volume mounting.
# If you need custom extensions or tools, you can extend the image here.
FROM postgres:16

# Example: install postgis or other extensions if needed
# RUN apt-get update && apt-get install -y postgresql-15-postgis-3

# Copy init scripts (already handled by docker-compose volume)
# COPY ./dev_tools/scripts/postgres-init /docker-entrypoint-initdb.d/

# No further customization needed for basic setup
