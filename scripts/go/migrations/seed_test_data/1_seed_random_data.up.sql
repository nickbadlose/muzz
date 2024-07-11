INSERT INTO public."user" (email, password, name, gender, age, location)
SELECT
    'email-' || round(random()*100000000) || '-' ||  round(random()*100000000) || '-' ||  round(random()*100000000) || '@email.com',
    '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK',
    'name',
    (array['male','female','unspecified'])[floor(random() * 3 + 1)],
    floor(random()*(50-18+1))+18,
    ST_SetSRID(ST_MakePoint((random()*360.0) - 180.0,(acos(1.0 - 2.0 * random()) * 2.0 - pi()) * 90.0 / pi()),4326)
FROM generate_series(1,1000) vtab ON CONFLICT DO NOTHING;

DO $FN$
    DECLARE rand int;
        DECLARE rows int = 1000;
    BEGIN
        FOR counter IN 1..rows LOOP
                FOR rowInsertCounter in 0..19 LOOP
                        rand = floor(random() * (((rowInsertCounter+1)*(rows/20)) - (rowInsertCounter*(rows/20)) + 1))+(rowInsertCounter*(rows/20));
                        INSERT INTO public.swipe (user_id, swiped_user_id, preference)
                        VALUES
                            (
                                counter,
                                CASE WHEN rand = counter THEN (rand-1) ELSE rand END,
                                (array[true, false])[floor(random() * 2 + 1)]
                            ) ON CONFLICT (user_id, swiped_user_id) DO NOTHING;
                    END LOOP;
            END LOOP;
    END;
$FN$;