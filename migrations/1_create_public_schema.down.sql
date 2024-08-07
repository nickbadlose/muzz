-- Pseudo file to prevent down migrations from failing,
-- as we can't delete the public schema whilst the migrations table exists.
-- Due to this issue, we need to manually delete the migrations table and public schema.
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS postgis;