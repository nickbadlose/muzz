CREATE TABLE public.swipe (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    swiped_user_id INT NOT NULL,
    preference BOOLEAN NOT NULL,

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    CONSTRAINT fk_swiped_user_id FOREIGN KEY (user_id) REFERENCES public.user (id) ON DELETE CASCADE,
    CONSTRAINT unique_swiped_user_per_user UNIQUE (user_id, swiped_user_id),
    CONSTRAINT no_matching_user_ids CHECK (user_id <> swiped_user_id)
);