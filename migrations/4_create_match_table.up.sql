CREATE TABLE public.match (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    matched_user_id INT NOT NULL,

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    CONSTRAINT fk_matched_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    -- no duplicate matches.
    CONSTRAINT unique_matched_user_per_user UNIQUE (user_id, matched_user_id),
    -- user cannot match themselves.
    CONSTRAINT no_matching_user_ids CHECK (user_id <> matched_user_id)
);