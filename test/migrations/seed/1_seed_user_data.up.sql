INSERT INTO public.user (email, password, name, gender, age, location)
VALUES
    -- password is Pa55w0rd! encrypted.
    ('test@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test', 'male', 25, '0101000020E61000002EFF21FDF63514C0355EBA490C224940'),
    ('test2@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test2', 'female', 18, '0101000020E61000002EFF21FDF63514C0355EBA490C224940'),
    ('test3@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test3', 'female', 28, '0101000020E6100000E4839ECDAACF0240F6285C8FC26D4840'),
    ('test4@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test4', 'female', 40, '0101000020E6100000DA1B7C613255C0BFFE43FAEDEBC04940'),
    ('test5@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test5', 'female', 25, '0101000020E610000074B515FBCBEE07C0DCD7817346B44A40'),
    ('test6@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test6', 'male', 29, '0101000020E6100000A913D044D8F001C05AF5B9DA8ABD4A40'),
    ('test7@test.com', '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK', 'test7', 'unspecified', 27, '0101000020E61000003480B74082E2F9BFA1D634EF387D4B40');