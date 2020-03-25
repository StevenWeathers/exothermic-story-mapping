--
-- Extensions
--
CREATE extension IF NOT EXISTS "uuid-ossp";

--
-- Tables
--
CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(64),
    created_date TIMESTAMP DEFAULT NOW(),
    last_active TIMESTAMP DEFAULT NOW(),
    email VARCHAR(320) UNIQUE,
    password TEXT,
    type VARCHAR(128) DEFAULT 'GUEST'
);

CREATE TABLE IF NOT EXISTS storyboard (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(256),
    owner_id UUID REFERENCES users NOT NULL,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS storyboard_user (
    storyboard_id UUID REFERENCES storyboard NOT NULL,
    user_id UUID REFERENCES users NOT NULL,
    active BOOL DEFAULT false,
    PRIMARY KEY (storyboard_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_reset (
    reset_id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES users NOT NULL,
    created_date TIMESTAMP DEFAULT NOW(),
    expire_date TIMESTAMP DEFAULT NOW() + INTERVAL '1 hour'
);

--
-- Table Alterations
--

--
-- Stored Procedures
--

-- Reset All Users to Inactive, used by server restart --
CREATE OR REPLACE PROCEDURE deactivate_all_users()
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_user SET active = false WHERE active = true;
END;
$$;

-- Set Storyboard Owner --
CREATE OR REPLACE PROCEDURE set_storyboard_owner(storyboardId UUID, ownerId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard SET updated_date = NOW(), owner_id = ownerId WHERE id = storyboardId;
END;
$$;

-- Delete Storyboard --
CREATE OR REPLACE PROCEDURE delete_storyboard(storyboardId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM storyboard_user WHERE storyboard_id = storyboardId;
    DELETE FROM storyboard WHERE id = storyboardId;

    COMMIT;
END;
$$;

-- Reset User Password --
CREATE OR REPLACE PROCEDURE reset_user_password(resetId UUID, userPassword TEXT)
LANGUAGE plpgsql AS $$
DECLARE matchedUserId UUID;
BEGIN
	matchedUserId := (
        SELECT w.id
        FROM user_reset wr
        LEFT JOIN user w ON w.id = wr.user_id
        WHERE wr.reset_id = resetId AND NOW() < wr.expire_date
    );

    IF matchedUserId IS NULL THEN
        -- attempt delete incase reset record expired
        DELETE FROM user_reset WHERE reset_id = resetId;
        RAISE 'Valid Reset ID not found';
    END IF;

    UPDATE users SET password = userPassword WHERE id = matchedUserId;
    DELETE FROM user_reset WHERE reset_id = resetId;

    COMMIT;
END;
$$;
