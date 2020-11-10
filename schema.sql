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

CREATE TABLE IF NOT EXISTS storyboard_goal (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    storyboard_id UUID REFERENCES storyboard NOT NULL,
    name VARCHAR(256),
    sort_order INTEGER,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(storyboard_id, sort_order)
);

CREATE TABLE IF NOT EXISTS storyboard_column (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    storyboard_id UUID REFERENCES storyboard NOT NULL,
    goal_id UUID REFERENCES storyboard_goal NOT NULL,
    name VARCHAR(256),
    sort_order INTEGER,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(goal_id, sort_order)
);

CREATE TABLE IF NOT EXISTS storyboard_story (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    storyboard_id UUID REFERENCES storyboard NOT NULL,
    goal_id UUID REFERENCES storyboard_goal NOT NULL,
    column_id UUID REFERENCES storyboard_column NOT NULL,
    name VARCHAR(256),
    color VARCHAR(32) DEFAULT 'blue',
    content TEXT,
    sort_order INTEGER,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(column_id, sort_order)
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

CREATE TABLE IF NOT EXISTS user_verify (
    verify_id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID REFERENCES users NOT NULL,
    created_date TIMESTAMP DEFAULT NOW(),
    expire_date TIMESTAMP DEFAULT NOW() + INTERVAL '24 hour'
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT NOT NULL PRIMARY KEY,
    user_id UUID REFERENCES users NOT NULL,
    name VARCHAR(256) NOT NULL,
    active BOOL DEFAULT true,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, name)
);

--
-- Table Alterations
--
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified BOOL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(128) DEFAULT 'identicon';
ALTER TABLE storyboard_user ADD COLUMN IF NOT EXISTS abandoned BOOL DEFAULT false;

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
    DELETE FROM storyboard_story WHERE storyboard_id = storyboardId;
    DELETE FROM storyboard_column WHERE storyboard_id = storyboardId;
    DELETE FROM storyboard_goal WHERE storyboard_id = storyboardId;
    DELETE FROM storyboard_user WHERE storyboard_id = storyboardId;
    DELETE FROM storyboard WHERE id = storyboardId;

    COMMIT;
END;
$$;

-- Create a Storyboard Goal --
CREATE OR REPLACE PROCEDURE create_storyboard_goal(storyBoardId UUID, goalName VARCHAR(256))
LANGUAGE plpgsql AS $$
DECLARE sortOrder INTEGER;
BEGIN
    sortOrder := (SELECT coalesce(MAX(sort_order), 0) FROM storyboard_goal WHERE storyboard_id = storyBoardId) + 1;
    INSERT INTO
        storyboard_goal
        (storyboard_id, sort_order, name)
        VALUES (storyBoardId, sortOrder, goalName);
END;
$$;

-- Revise a Storyboard Goal --
CREATE OR REPLACE PROCEDURE update_storyboard_goal(goalId UUID, goalName VARCHAR(256))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_goal SET name = goalName, updated_date = NOW() WHERE id = goalId;
END;
$$;

-- Delete a Storyboard Goal --
CREATE OR REPLACE PROCEDURE delete_storyboard_goal(goalId UUID)
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
DECLARE sortOrder INTEGER;
BEGIN
    SELECT sort_order, storyboard_id INTO sortOrder, storyboardId FROM storyboard_goal WHERE id = goalId;

    DELETE FROM storyboard_story WHERE goal_id = goalId;
    DELETE FROM storyboard_column WHERE goal_id = goalId;
    DELETE FROM storyboard_goal WHERE id = goalId;
    UPDATE storyboard_goal sg SET sort_order = (sg.sort_order - 1) WHERE sg.storyboard_id = storyBoardId AND sg.sort_order > sortOrder;
    
    COMMIT;
END;
$$;

-- Create a Storyboard Column --
CREATE OR REPLACE PROCEDURE create_storyboard_column(storyBoardId UUID, goalId UUID)
LANGUAGE plpgsql AS $$
DECLARE sortOrder INTEGER;
BEGIN
    sortOrder := (SELECT coalesce(MAX(sort_order), 0) FROM storyboard_column WHERE goal_id = goalId) + 1;
    INSERT INTO storyboard_column (storyboard_id, goal_id, sort_order) VALUES (storyBoardId, goalId, sortOrder);
END;
$$;

-- Delete a Storyboard Column --
CREATE OR REPLACE PROCEDURE delete_storyboard_column(columnId UUID)
LANGUAGE plpgsql AS $$
DECLARE goalId UUID;
DECLARE sortOrder INTEGER;
BEGIN
    SELECT goal_id, sort_order INTO goalId, sortOrder FROM storyboard_column WHERE id = columnId;

    DELETE FROM storyboard_story WHERE column_id = columnId;
    DELETE FROM storyboard_column WHERE id = columnId;
    UPDATE storyboard_column sc SET sort_order = (sc.sort_order - 1) WHERE sc.goal_id = goalId AND sc.sort_order > sortOrder;
    
    COMMIT;
END;
$$;

-- Create a Storyboard Story --
CREATE OR REPLACE PROCEDURE create_storyboard_story(storyBoardId UUID, goalId UUID, columnId UUID)
LANGUAGE plpgsql AS $$
DECLARE sortOrder INTEGER;
BEGIN
    sortOrder := (SELECT coalesce(MAX(sort_order), 0) FROM storyboard_story WHERE columnId = columnId) + 1;
    INSERT INTO storyboard_story (storyboard_id, goal_id, column_id, sort_order) VALUES (storyBoardId, goalId, columnId, sortOrder);
END;
$$;

-- Revise a Storyboard Story Name --
CREATE OR REPLACE PROCEDURE update_story_name(storyId UUID, storyName VARCHAR(256))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_story SET name = storyName, updated_date = NOW() WHERE id = storyId;
END;
$$;

-- Revise a Storyboard Story Content --
CREATE OR REPLACE PROCEDURE update_story_content(storyId UUID, storyContent TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_story SET content = storyContent, updated_date = NOW() WHERE id = storyId;
END;
$$;

-- Revise a Storyboard Story Color --
CREATE OR REPLACE PROCEDURE update_story_color(storyId UUID, storyColor VARCHAR(32))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_story SET color = storyColor, updated_date = NOW() WHERE id = storyId;
END;
$$;

-- Move a Storyboard Story to a new column and/or goal --
CREATE OR REPLACE PROCEDURE move_story(storyId UUID, goalId UUID, columnId UUID, placeBefore TEXT)
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
DECLARE srcGoalId UUID;
DECLARE srcColumnId UUID;
DECLARE srcSortOrder INTEGER;
DECLARE storyName VARCHAR(256);
DECLARE storyColor VARCHAR(32);
DECLARE storyContent TEXT;
DECLARE createdDate TIMESTAMP;
DECLARE targetSortOrder INTEGER;
BEGIN
    -- Get Story current details
    SELECT 
        storyboard_id, goal_id, column_id, sort_order, name, color, content, created_date
    INTO
        storyboardId, srcGoalId, srcColumnId, srcSortOrder, storyName, storyColor, storyContent, createdDate
    FROM storyboard_story WHERE id = storyId;

    -- Get target sort order
    IF placeBefore = '' THEN
        SELECT coalesce(max(sort_order), 0) + 1 INTO targetSortOrder FROM storyboard_story WHERE column_id = columnId;
    ELSE
        SELECT sort_order INTO targetSortOrder FROM storyboard_story WHERE column_id = columnId AND id = placeBefore::UUID;
    END IF;

    -- Remove from source column
    DELETE FROM storyboard_story WHERE id = storyId;
    -- Update sort order in src column
    UPDATE storyboard_story ss SET sort_order = (t.sort_order - 1)
    FROM (
        SELECT id, sort_order FROM storyboard_story
        WHERE column_id = srcColumnId AND sort_order > srcSortOrder
        ORDER BY sort_order ASC
        FOR UPDATE
    ) AS t
    WHERE ss.id = t.id;

    -- Update sort order for any story that should come after newly moved story
    UPDATE storyboard_story ss SET sort_order = (t.sort_order + 1)
    FROM (
        SELECT id, sort_order FROM storyboard_story
        WHERE column_id = columnId AND sort_order >= targetSortOrder
        ORDER BY sort_order DESC
        FOR UPDATE
    ) AS t
    WHERE ss.id = t.id;

    -- Finally, insert new story in its ordered place
    INSERT INTO
        storyboard_story (
            storyboard_id, goal_id, column_id, sort_order, name, color, content, created_date
        )
    VALUES (
        storyBoardId, goalId, columnId, targetSortOrder, storyName, storyColor, storyContent, createdDate
    );

    COMMIT;
END;
$$;

-- Delete a Storyboard Story --
CREATE OR REPLACE PROCEDURE delete_storyboard_story(storyId UUID)
LANGUAGE plpgsql AS $$
DECLARE columnId UUID;
DECLARE sortOrder INTEGER;
BEGIN
    SELECT column_id, sort_order INTO columnId, sortOrder FROM storyboard_story WHERE id = storyId;
    DELETE FROM storyboard_story WHERE id = storyId;
    UPDATE storyboard_story ss SET sort_order = (ss.sort_order - 1) WHERE ss.column_id = columnId AND ss.sort_order > sortOrder;
    
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

    UPDATE users SET password = userPassword, last_active = NOW() WHERE id = matchedUserId;
    DELETE FROM user_reset WHERE reset_id = resetId;

    COMMIT;
END;
$$;

-- Update User Password --
CREATE OR REPLACE PROCEDURE update_user_password(userId UUID, userPassword TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET password = userPassword, last_active = NOW() WHERE id = userId;

    COMMIT;
END;
$$;

-- Verify a user account email
CREATE OR REPLACE PROCEDURE verify_user_account(verifyId UUID)
LANGUAGE plpgsql AS $$
DECLARE matchedUserId UUID;
BEGIN
	matchedUserId := (
        SELECT usr.id
        FROM user_verify uv
        LEFT JOIN users usr ON usr.id = uv.user_id
        WHERE uv.verify_id = verifyId AND NOW() < uv.expire_date
    );

    IF matchedUserId IS NULL THEN
        -- attempt delete incase verify record expired
        DELETE FROM user_verify WHERE verify_id = verifyId;
        RAISE 'Valid Verify ID not found';
    END IF;

    UPDATE users SET verified = 'TRUE', last_active = NOW() WHERE id = matchedUserId;
    DELETE FROM user_verify WHERE verify_id = verifyId;

    COMMIT;
END;
$$;

-- Promote User to ADMIN by ID --
CREATE OR REPLACE PROCEDURE promote_user(userId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'ADMIN' WHERE id = userId;

    COMMIT;
END;
$$;

-- Promote User to ADMIN by Email --
CREATE OR REPLACE PROCEDURE promote_user_by_email(userEmail VARCHAR(320))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'ADMIN' WHERE email = userEmail;

    COMMIT;
END;
$$;

-- Demote User to Registered by ID --
CREATE OR REPLACE PROCEDURE demote_user(userId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'REGISTERED' WHERE id = userId;

    COMMIT;
END;
$$;

--
-- Stored Functions
--

-- Create a Storyboard
DROP FUNCTION IF EXISTS create_storyboard(UUID, VARCHAR);
CREATE FUNCTION create_storyboard(ownerId UUID, storyboardName VARCHAR(256)) RETURNS UUID 
AS $$ 
DECLARE storyId UUID;
BEGIN
    INSERT INTO storyboard (owner_id, name) VALUES (ownerId, storyboardName) RETURNING id INTO storyId;

    RETURN storyId;
END;
$$ LANGUAGE plpgsql;

-- Get Storyboards by User ID
DROP FUNCTION IF EXISTS get_storyboards_by_user(uuid);
CREATE FUNCTION get_storyboards_by_user(userId UUID) RETURNS table (
    id UUID, name VARCHAR(256), owner_id UUID
) AS $$
BEGIN
    RETURN QUERY
        SELECT b.id, b.name, b.owner_id
		FROM storyboard b
		LEFT JOIN storyboard_user bw ON b.id = bw.storyboard_id WHERE bw.user_id = userId AND bw.abandoned = false
		GROUP BY b.id ORDER BY b.created_date DESC;
END;
$$ LANGUAGE plpgsql;

-- Get a Storyboards Goals --
DROP FUNCTION IF EXISTS get_storyboard_goals(uuid);
CREATE FUNCTION get_storyboard_goals(storyboardId UUID) RETURNS table (
    id UUID, sort_order INTEGER, name VARCHAR(256), columns JSON
) AS $$
BEGIN
    RETURN QUERY
        SELECT
            sg.id,
            sg.sort_order,
            sg.name,
            COALESCE(json_agg(to_jsonb(t) - 'goal_id' ORDER BY t.sort_order) FILTER (WHERE t.id IS NOT NULL), '[]') AS columns           
        FROM storyboard_goal sg
        LEFT JOIN (
            SELECT
                sc.*,
                COALESCE(
                    json_agg(ss ORDER BY ss.sort_order) FILTER (WHERE ss.id IS NOT NULL), '[]'
                ) AS stories
            FROM storyboard_column sc
            LEFT JOIN storyboard_story ss ON ss.column_id = sc.id
            GROUP BY sc.id
        ) t ON t.goal_id = sg.id
        WHERE sg.storyboard_id = storyboardId
        GROUP BY sg.id
        ORDER BY sg.sort_order;
END;
$$ LANGUAGE plpgsql;

-- Get a User by ID
DROP FUNCTION IF EXISTS get_user(UUID);
CREATE FUNCTION get_user(userId UUID) RETURNS table (
    id UUID, name VARCHAR(64), email VARCHAR(320), type VARCHAR(128), verified BOOL
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, coalesce(u.email, ''), u.type, u.verified FROM users u WHERE u.id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get Storyboard Users
DROP FUNCTION IF EXISTS get_storyboard_users(uuid);
CREATE FUNCTION get_storyboard_users(storyboardId UUID) RETURNS table (
    id UUID, name VARCHAR(256), active BOOL
) AS $$
BEGIN
    RETURN QUERY
        SELECT
			w.id, w.name, bw.active
		FROM storyboard_user bw
		LEFT JOIN users w ON bw.user_id = w.id
		WHERE bw.storyboard_id = storyboardId
		ORDER BY w.name;
END;
$$ LANGUAGE plpgsql;

-- Get Storyboard User by id
DROP FUNCTION IF EXISTS get_storyboard_user(uuid, uuid);
CREATE FUNCTION get_storyboard_user(storyboardId UUID, userId UUID) RETURNS table (
    id UUID, name VARCHAR(256), active BOOL
) AS $$
BEGIN
    RETURN QUERY
        SELECT
			w.id, w.name, coalesce(bw.active, FALSE)
		FROM users w
		LEFT JOIN storyboard_user bw ON bw.user_id = w.id AND bw.storyboard_id = storyboardId
		WHERE w.id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get User Auth by Email
DROP FUNCTION IF EXISTS get_user_auth_by_email(VARCHAR);
CREATE FUNCTION get_user_auth_by_email(userEmail VARCHAR(320)) RETURNS table (
    id UUID, name VARCHAR(64), email VARCHAR(320), type VARCHAR(128), password TEXT
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, coalesce(u.email, ''), u.type, u.password FROM users u WHERE u.email = userEmail;
END;
$$ LANGUAGE plpgsql;

-- Get Application Stats e.g. total user and storyboard counts
DROP FUNCTION IF EXISTS get_app_stats();
CREATE FUNCTION get_app_stats(
    OUT unregistered_user_count INTEGER,
    OUT registered_user_count INTEGER,
    OUT storyboard_count INTEGER
) AS $$
BEGIN
    SELECT COUNT(*) INTO unregistered_user_count FROM users WHERE email IS NULL;
    SELECT COUNT(*) INTO registered_user_count FROM users WHERE email IS NOT NULL;
    SELECT COUNT(*) INTO storyboard_count FROM storyboard;
END;
$$ LANGUAGE plpgsql;

-- Insert a new user password reset
DROP FUNCTION IF EXISTS insert_user_reset(VARCHAR);
CREATE FUNCTION insert_user_reset(
    IN userEmail VARCHAR(320),
    OUT resetId UUID,
    OUT userId UUID,
    OUT userName VARCHAR(64)
)
AS $$ 
BEGIN
    SELECT id, name INTO userId, userName FROM users WHERE email = userEmail;
    INSERT INTO user_reset (user_id) VALUES (userId) RETURNING reset_id INTO resetId;
END;
$$ LANGUAGE plpgsql;

-- Register a new user
DROP FUNCTION IF EXISTS register_user(VARCHAR, VARCHAR, TEXT, VARCHAR);
CREATE FUNCTION register_user(
    IN userName VARCHAR(64),
    IN userEmail VARCHAR(320),
    IN hashedPassword TEXT,
    IN userType VARCHAR(128),
    OUT userId UUID,
    OUT verifyId UUID
)
AS $$
BEGIN
    INSERT INTO users (name, email, password, type)
    VALUES (userName, userEmail, hashedPassword, userType)
    RETURNING id INTO userId;

    INSERT INTO user_verify (user_id) VALUES (userId) RETURNING verify_id INTO verifyId;
END;
$$ LANGUAGE plpgsql;

-- Register a new user from existing GUEST
DROP FUNCTION IF EXISTS register_existing_user(UUID, VARCHAR, VARCHAR, TEXT, VARCHAR);
CREATE FUNCTION register_existing_user(
    IN activeUserId UUID,
    IN userName VARCHAR(64),
    IN userEmail VARCHAR(320),
    IN hashedPassword TEXT,
    IN userType VARCHAR(128),
    OUT userId UUID,
    OUT verifyId UUID
)
AS $$
BEGIN
    UPDATE users
    SET
         name = userName,
         email = userEmail,
         password = hashedPassword,
         type = userType,
         last_active = NOW()
    WHERE id = activeUserId
    RETURNING id INTO userId;

    INSERT INTO user_verify (user_id) VALUES (userId) RETURNING verify_id INTO verifyId;
END;
$$ LANGUAGE plpgsql;