CREATE TABLE IF NOT EXISTS public."User" (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    gender TEXT NOT NULL,
    age INT NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    unique(email)
);