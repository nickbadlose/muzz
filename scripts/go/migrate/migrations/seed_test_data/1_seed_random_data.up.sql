INSERT INTO public."user" (email, password, name, gender, age, location)
SELECT
    'email-' || round(random()*100000000) || '-' ||  round(random()*100000000) || '-' ||  round(random()*100000000) || '@email.com',
    '$2a$06$ewczVCXHOOgz2K0AdtTDauqMMhoUAcQu2AOng0CZdOrgu4QgHFpLK',
    'name',
    (array['male','female','unspecified'])[floor(random() * 3 + 1)],
    floor(random()*(50-18+1))+18,
    ST_SetSRID(ST_MakePoint((random()*360.0) - 180.0,(acos(1.0 - 2.0 * random()) * 2.0 - pi()) * 90.0 / pi()),4326)
FROM generate_series(1,10000) vtab ON CONFLICT DO NOTHING;

DO $FN$
    -- random swiped_user_id to insert into swipe table
    DECLARE rand int;
    -- backup random integers in case rand matches the current user_id in the swipe insert.
    DECLARE secondary_rand int;
        -- users is the number of users to seed data for
        DECLARE users int = 10000;
    BEGIN
        -- 5% of users will be simulated as inactive
        FOR counter IN 1..(users-(users / 20)) LOOP
                -- add 20 swipes per user
                FOR rowInsertCounter in 0..19 LOOP
                        rand = floor(random() * (((rowInsertCounter+1)*(users/20)) - (rowInsertCounter*(users/20)) + 1))+(rowInsertCounter*(users/20));
                        secondary_rand = floor(random() * (((rowInsertCounter+1)*(users/20)) - (rowInsertCounter*(users/20)) + 1))+(rowInsertCounter*(users/20));
                        INSERT INTO public.swipe (user_id, swiped_user_id, preference)
                        VALUES
                            (
                                counter,
                                -- if user_id and swiped_user_id are equal, use backup swiped_user_id,
                                -- if that is still equal subtract 1 from the value and insert.
                                -- this prevents 'no_matching_user_ids' constraint killing script.
                                CASE WHEN rand = counter THEN (CASE WHEN secondary_rand = counter THEN (rand-1) ELSE secondary_rand END) ELSE rand END,
                                (array[true, false])[floor(random() * 2 + 1)]
                            -- if row is a duplicate insert, it will be skipped instead of stopping the whole batch
                            ) ON CONFLICT (user_id, swiped_user_id) DO NOTHING;
                    END LOOP;
            END LOOP;
    END;
$FN$;