INSERT INTO public.user (email, password, name, gender, age, location)
VALUES
    -- password is Pa55w0rd! encrypted.
    ('test@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test', 'male', 25, '0101000020E61000002EFF21FDF63514C0355EBA490C224940'),
    ('test2@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test2', 'female', 18, '0101000020E61000002EFF21FDF63514C0355EBA490C224940');

INSERT INTO public.swipe (user_id, swiped_user_id, preference)
VALUES
    (1, 2, true),
    (1, 2, false);