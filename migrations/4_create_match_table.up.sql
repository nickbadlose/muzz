CREATE TABLE public.match (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    matched_user_id INT NOT NULL,
--  TODO
--   check this and add tests to check constraints. isolate them and run them using golang migrate, check error returned from migrator.
--   check for a clean way to insert only one row into match table
--   document migrations if necessary, especially checks. Tests may self document constraints if we name tests after constraints.
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    CONSTRAINT fk_matched_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    CONSTRAINT unique_matched_user_per_user UNIQUE (user_id, matched_user_id),
    CONSTRAINT no_matching_user_ids CHECK (user_id <> matched_user_id)
);