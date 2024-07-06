ALTER TABLE public.match DROP CONSTRAINT fk_user_id;
ALTER TABLE public.match DROP CONSTRAINT fk_matched_user_id;

DROP TABLE public.match;