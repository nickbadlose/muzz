ALTER TABLE public.swipe DROP CONSTRAINT fk_user_id;
ALTER TABLE public.swipe DROP CONSTRAINT fk_swiped_user_id;

DROP TABLE public.swipe;