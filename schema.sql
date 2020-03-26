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

    UPDATE users SET password = userPassword WHERE id = matchedUserId;
    DELETE FROM user_reset WHERE reset_id = resetId;

    COMMIT;
END;
$$;
