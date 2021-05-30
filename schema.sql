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

CREATE TABLE IF NOT EXISTS storyboard_persona (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    storyboard_id UUID NOT NULL,
    name VARCHAR(256) NOT NULL,
    role VARCHAR(256),
    description TEXT,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(storyboard_id, name),
    CONSTRAINT sp_storyboard_id FOREIGN KEY(storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS story_comment (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    storyboard_id UUID NOT NULL,
    story_id UUID NOT NULL,
    comment TEXT,
    user_id UUID NOT NULL,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    CONSTRAINT stc_storyboard_id FOREIGN KEY(storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE,
    CONSTRAINT stc_story_id FOREIGN KEY(story_id) REFERENCES storyboard_story(id) ON DELETE CASCADE,
    CONSTRAINT stc_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(256),
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organization_user (
    organization_id UUID,
    user_id UUID,
    role VARCHAR(16) NOT NULL DEFAULT 'MEMBER',
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT ou_organization_id FOREIGN KEY(organization_id) REFERENCES organization(id) ON DELETE CASCADE,
    CONSTRAINT ou_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization_department (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    organization_id UUID,
    name VARCHAR(256),
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    UNIQUE(organization_id, name),
    CONSTRAINT od_organization_id FOREIGN KEY(organization_id) REFERENCES organization(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS department_user (
    department_id UUID,
    user_id UUID,
    role VARCHAR(16) NOT NULL DEFAULT 'MEMBER',
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (department_id, user_id),
    CONSTRAINT du_department_id FOREIGN KEY(department_id) REFERENCES organization_department(id) ON DELETE CASCADE,
    CONSTRAINT du_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(256),
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_user (
    team_id UUID,
    user_id UUID,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    role VARCHAR(16) NOT NULL DEFAULT 'MEMBER',
    PRIMARY KEY (team_id, user_id),
    CONSTRAINT tu_team_id FOREIGN KEY(team_id) REFERENCES team(id) ON DELETE CASCADE,
    CONSTRAINT tu_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization_team (
    organization_id UUID,
    team_id UUID,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (organization_id, team_id),
    UNIQUE(team_id),
    CONSTRAINT ot_organization_id FOREIGN KEY(organization_id) REFERENCES organization(id) ON DELETE CASCADE,
    CONSTRAINT ot_team_id FOREIGN KEY(team_id) REFERENCES team(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS department_team (
    department_id UUID,
    team_id UUID,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (department_id, team_id),
    UNIQUE(team_id),
    CONSTRAINT dt_department_id FOREIGN KEY(department_id) REFERENCES organization_department(id) ON DELETE CASCADE,
    CONSTRAINT dt_team_id FOREIGN KEY(team_id) REFERENCES team(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team_storyboard (
    team_id UUID,
    storyboard_id UUID,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (team_id, storyboard_id),
    CONSTRAINT tb_team_id FOREIGN KEY(team_id) REFERENCES team(id) ON DELETE CASCADE,
    CONSTRAINT tb_storyboard_id FOREIGN KEY(storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alert (
    id UUID  NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    type VARCHAR(128) DEFAULT 'NEW',
    content TEXT NOT NULL,
    active BOOLEAN DEFAULT true,
    allow_dismiss BOOLEAN DEFAULT true,
    registered_only BOOLEAN DEFAULT true,
    created_date TIMESTAMP DEFAULT NOW(),
    updated_date TIMESTAMP DEFAULT NOW()
);

--
-- Table Alterations
--
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified BOOL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(128) DEFAULT 'identicon';
ALTER TABLE users ADD COLUMN IF NOT EXISTS country VARCHAR(2);
ALTER TABLE users ADD COLUMN IF NOT EXISTS company VARCHAR(256);
ALTER TABLE users ADD COLUMN IF NOT EXISTS job_title VARCHAR(128);
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_date TIMESTAMP DEFAULT NOW();

ALTER TABLE storyboard_user ADD COLUMN IF NOT EXISTS abandoned BOOL DEFAULT false;
ALTER TABLE storyboard_story ADD COLUMN IF NOT EXISTS points INTEGER;
ALTER TABLE storyboard_story ADD COLUMN IF NOT EXISTS closed BOOL DEFAULT false;
ALTER TABLE storyboard_story ALTER COLUMN color SET DEFAULT 'gray';
ALTER TABLE storyboard ADD COLUMN IF NOT EXISTS color_legend JSONB DEFAULT '[{"color":"gray","legend":""},{"color":"red","legend":""},{"color":"orange","legend":""},{"color":"yellow","legend":""},{"color":"green","legend":""},{"color":"teal","legend":""},{"color":"blue","legend":""},{"color":"indigo","legend":""},{"color":"purple","legend":""},{"color":"pink","legend":""}]'::JSONB;

DO $$
BEGIN
    --
    -- Constraints
    --
    ALTER TABLE storyboard DROP CONSTRAINT IF EXISTS storyboard_owner_id_fkey;
    ALTER TABLE storyboard_user DROP CONSTRAINT IF EXISTS storyboard_user_storyboard_id_fkey;
    ALTER TABLE storyboard_user DROP CONSTRAINT IF EXISTS storyboard_user_user_id_fkey;
    ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_user_id_fkey;
    ALTER TABLE user_verify DROP CONSTRAINT IF EXISTS user_verify_user_id_fkey;
    ALTER TABLE user_reset DROP CONSTRAINT IF EXISTS user_reset_user_id_fkey;
    ALTER TABLE storyboard_goal DROP CONSTRAINT IF EXISTS storyboard_goal_storyboard_id_fkey;
    ALTER TABLE storyboard_column DROP CONSTRAINT IF EXISTS storyboard_column_storyboard_id_fkey;
    ALTER TABLE storyboard_column DROP CONSTRAINT IF EXISTS storyboard_column_goal_id_fkey;
    ALTER TABLE storyboard_story DROP CONSTRAINT IF EXISTS storyboard_story_storyboard_id_fkey;
    ALTER TABLE storyboard_story DROP CONSTRAINT IF EXISTS storyboard_story_goal_id_fkey;
    ALTER TABLE storyboard_story DROP CONSTRAINT IF EXISTS storyboard_story_column_id_fkey;

    BEGIN
        ALTER TABLE storyboard ADD CONSTRAINT s_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard constraint s_owner_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_user ADD CONSTRAINT su_storyboard_id_fkey FOREIGN KEY (storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_user constraint su_storyboard_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_user ADD CONSTRAINT su_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_user constraint su_user_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE user_reset ADD CONSTRAINT ur_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'user_reset constraint ur_user_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE user_verify ADD CONSTRAINT uv_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'user_verify constraint uv_user_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE api_keys ADD CONSTRAINT apk_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'api_keys constraint apk_user_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_goal ADD CONSTRAINT sg_storyboard_id_fkey FOREIGN KEY (storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_goal constraint sg_storyboard_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_column ADD CONSTRAINT sc_storyboard_id_fkey FOREIGN KEY (storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_column constraint sc_storyboard_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_column ADD CONSTRAINT sc_goal_id_fkey FOREIGN KEY (goal_id) REFERENCES storyboard_goal(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_column constraint sc_goal_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_story ADD CONSTRAINT ss_storyboard_id_fkey FOREIGN KEY (storyboard_id) REFERENCES storyboard(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_story constraint ss_storyboard_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_story ADD CONSTRAINT ss_goal_id_fkey FOREIGN KEY (goal_id) REFERENCES storyboard_goal(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_story constraint ss_goal_id_fkey already exists';
    END;

    BEGIN
        ALTER TABLE storyboard_story ADD CONSTRAINT ss_column_id_fkey FOREIGN KEY (column_id) REFERENCES storyboard_column(id) ON DELETE CASCADE;
        EXCEPTION
        WHEN duplicate_object THEN RAISE NOTICE 'storyboard_story constraint ss_column_id_fkey already exists';
    END;
END $$;

--
-- Views
--
CREATE MATERIALIZED VIEW IF NOT EXISTS active_countries AS SELECT DISTINCT country FROM users;

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

-- Revise Storyboard ColorLegend --
CREATE OR REPLACE PROCEDURE revise_color_legend(storyboardId UUID, colorLegend JSONB)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard SET updated_date = NOW(), color_legend = colorLegend WHERE id = storyboardId;
END;
$$;

-- Delete Storyboard --
CREATE OR REPLACE PROCEDURE delete_storyboard(storyboardId UUID)
LANGUAGE plpgsql AS $$
BEGIN
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

    UPDATE storyboard SET updated_date = NOW() WHERE id = storyBoardId;
END;
$$;

-- Revise a Storyboard Goal --
CREATE OR REPLACE PROCEDURE update_storyboard_goal(goalId UUID, goalName VARCHAR(256))
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_goal SET name = goalName, updated_date = NOW() WHERE id = goalId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
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
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
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
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyBoardId;
END;
$$;

-- Revise a Storyboard Column --
CREATE OR REPLACE PROCEDURE revise_storyboard_column(storyBoardId UUID, columnId UUID, columnName VARCHAR(256))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_column SET name = columnName, updated_date = NOW() WHERE id = columnId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyBoardId;
END;
$$;

-- Delete a Storyboard Column --
CREATE OR REPLACE PROCEDURE delete_storyboard_column(columnId UUID)
LANGUAGE plpgsql AS $$
DECLARE goalId UUID;
DECLARE sortOrder INTEGER;
DECLARE storyboardId UUID;
BEGIN
    SELECT goal_id, sort_order INTO goalId, sortOrder FROM storyboard_column WHERE id = columnId;

    DELETE FROM storyboard_story WHERE column_id = columnId;
    DELETE FROM storyboard_column WHERE id = columnId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard_column sc SET sort_order = (sc.sort_order - 1) WHERE sc.goal_id = goalId AND sc.sort_order > sortOrder;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
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
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyBoardId;
END;
$$;

-- Revise a Storyboard Story Name --
CREATE OR REPLACE PROCEDURE update_story_name(storyId UUID, storyName VARCHAR(256))
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_story SET name = storyName, updated_date = NOW() WHERE id = storyId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
END;
$$;

-- Revise a Storyboard Story Content --
CREATE OR REPLACE PROCEDURE update_story_content(storyId UUID, storyContent TEXT)
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_story SET content = storyContent, updated_date = NOW() WHERE id = storyId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
END;
$$;

-- Revise a Storyboard Story Color --
CREATE OR REPLACE PROCEDURE update_story_color(storyId UUID, storyColor VARCHAR(32))
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_story SET color = storyColor, updated_date = NOW() WHERE id = storyId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
END;
$$;

-- Revise a Storyboard Story Points --
CREATE OR REPLACE PROCEDURE update_story_points(storyId UUID, updatedPoints INTEGER)
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_story SET points = updatedPoints, updated_date = NOW() WHERE id = storyId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
END;
$$;

-- Revise a Storyboard Story Closed status --
CREATE OR REPLACE PROCEDURE update_story_closed(storyId UUID, isClosed BOOL)
LANGUAGE plpgsql AS $$
DECLARE storyboardId UUID;
BEGIN
    UPDATE storyboard_story SET closed = isClosed, updated_date = NOW() WHERE id = storyId RETURNING storyboard_id INTO storyboardId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
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

    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;

    COMMIT;
END;
$$;

-- Delete a Storyboard Story --
CREATE OR REPLACE PROCEDURE delete_storyboard_story(storyId UUID)
LANGUAGE plpgsql AS $$
DECLARE columnId UUID;
DECLARE sortOrder INTEGER;
DECLARE storyboardId UUID;
BEGIN
    SELECT column_id, sort_order, storyboard_id INTO columnId, sortOrder, storyboardId FROM storyboard_story WHERE id = storyId;
    DELETE FROM storyboard_story WHERE id = storyId;
    UPDATE storyboard_story ss SET sort_order = (ss.sort_order - 1) WHERE ss.column_id = columnId AND ss.sort_order > sortOrder;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
    COMMIT;
END;
$$;

-- Add a comment to Storyboard Story --
CREATE OR REPLACE PROCEDURE story_comment_add(storyboardId UUID, storyId UUID, userId UUID, comment TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO story_comment (storyboard_id, story_id, user_id, comment) VALUES (storyboardId, storyId, userId, comment);
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
    COMMIT;
END;
$$;

-- -- Edit a comment on a Storyboard Story --
-- CREATE OR REPLACE PROCEDURE story_comment_edit(storyboardId UUID, commentId UUID, userId UUID, updatedComment TEXT)
-- LANGUAGE plpgsql AS $$
-- BEGIN
--     UPDATE story_comment SET comment = updatedComment, updated_date = NOW() WHERE id = commentId AND user_id = userId;
--     IF NOT found THEN
--         RAISE EXCEPTION 'Comment does not belong to user';
--     END IF;
--     UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
--     COMMIT;
-- END;
-- $$;

-- -- Delete a comment on a Storyboard Story --
-- CREATE OR REPLACE PROCEDURE story_comment_delete(storyboardId UUID, commentId UUID, userId UUID)
-- LANGUAGE plpgsql AS $$
-- BEGIN
--     DELETE FROM story_comment WHERE id = commentId AND user_id = userId;
--     IF NOT found THEN
--         RAISE EXCEPTION 'Comment does not belong to user';
--     END IF;
--     UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
--     COMMIT;
-- END;
-- $$;

-- Add a Persona to Storyboard --
CREATE OR REPLACE PROCEDURE persona_add(storyboardId UUID, personaName VARCHAR(256), personaRole VARCHAR(256), personaDescription TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO storyboard_persona (storyboard_id, name, role, description) VALUES (storyboardId, personaName, personaRole, personaDescription);
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
    COMMIT;
END;
$$;

-- Edit a Storyboard Persona --
CREATE OR REPLACE PROCEDURE persona_edit(storyboardId UUID, personaId UUID, personaName VARCHAR(256), personaRole VARCHAR(256), personaDescription TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE storyboard_persona SET name = personaName, role = personaRole, description = personaDescription, updated_date = NOW() WHERE id = personaId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
    COMMIT;
END;
$$;

-- Delete a Storyboard Persona --
CREATE OR REPLACE PROCEDURE persona_delete(storyboardId UUID, personaId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM storyboard_persona WHERE id = personaId;
    UPDATE storyboard SET updated_date = NOW() WHERE id = storyboardId;
    
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
        FROM user_reset ur
        LEFT JOIN user w ON w.id = ur.user_id
        WHERE ur.reset_id = resetId AND NOW() < ur.expire_date
    );

    IF matchedUserId IS NULL THEN
        -- attempt delete incase reset record expired
        DELETE FROM user_reset WHERE reset_id = resetId;
        RAISE 'Valid Reset ID not found';
    END IF;

    UPDATE users SET password = userPassword, last_active = NOW(), updated_date = NOW() WHERE id = matchedUserId;
    DELETE FROM user_reset WHERE reset_id = resetId;

    COMMIT;
END;
$$;

-- Update User Password --
CREATE OR REPLACE PROCEDURE update_user_password(userId UUID, userPassword TEXT)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET password = userPassword, last_active = NOW(), updated_date = NOW() WHERE id = userId;

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

    UPDATE users SET verified = 'TRUE', last_active = NOW(), updated_date = NOW() WHERE id = matchedUserId;
    DELETE FROM user_verify WHERE verify_id = verifyId;

    COMMIT;
END;
$$;

-- Promote User to ADMIN by ID --
CREATE OR REPLACE PROCEDURE promote_user(userId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'ADMIN', updated_date = NOW() WHERE id = userId;

    COMMIT;
END;
$$;

-- Promote User to ADMIN by Email --
CREATE OR REPLACE PROCEDURE promote_user_by_email(userEmail VARCHAR(320))
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'ADMIN', updated_date = NOW() WHERE email = userEmail;

    COMMIT;
END;
$$;

-- Demote User to Registered by ID --
CREATE OR REPLACE PROCEDURE demote_user(userId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET type = 'REGISTERED', updated_date = NOW() WHERE id = userId;

    COMMIT;
END;
$$;

-- Clean up Storyboards older than X Days --
CREATE OR REPLACE PROCEDURE clean_storyboards(daysOld INTEGER)
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM storyboard WHERE updated_date < (NOW() - daysOld * interval '1 day');

    COMMIT;
END;
$$;

-- Clean up Guest Users (and their created storyboards) older than X Days --
CREATE OR REPLACE PROCEDURE clean_guest_users(daysOld INTEGER)
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM users WHERE last_active < (NOW() - daysOld * interval '1 day') AND type = 'GUEST';
    REFRESH MATERIALIZED VIEW active_countries;

    COMMIT;
END;
$$;

-- Deletes a User and all his storyboard(s), api keys --
CREATE OR REPLACE PROCEDURE delete_user(userId UUID)
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM users WHERE id = userId;
    REFRESH MATERIALIZED VIEW active_countries;

    COMMIT;
END;
$$;

-- Updates a users profile --
CREATE OR REPLACE PROCEDURE user_profile_update(
    userId UUID,
    userName VARCHAR(64),
    userAvatar VARCHAR(128),
    userCountry VARCHAR(2),
    userCompany VARCHAR(256),
    userJobTitle VARCHAR(128)
)
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users
    SET name = userName, avatar = userAvatar, country = userCountry, company = userCompany, job_title = userJobTitle, last_active = NOW(), updated_date = NOW()
    WHERE id = userId;
    REFRESH MATERIALIZED VIEW active_countries;
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
		LEFT JOIN storyboard_user su ON b.id = su.storyboard_id WHERE su.user_id = userId AND su.abandoned = false
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
                    json_agg(stss ORDER BY stss.sort_order) FILTER (WHERE stss.id IS NOT NULL), '[]'
                ) AS stories
            FROM storyboard_column sc
            LEFT JOIN (
                SELECT
                    ss.*,
                    COALESCE(
                        json_agg(stcm ORDER BY stcm.created_date) FILTER (WHERE stcm.id IS NOT NULL), '[]'
                    ) AS comments
                FROM storyboard_story ss
                LEFT JOIN story_comment stcm ON stcm.story_id = ss.id
                GROUP BY ss.id
            ) stss ON stss.column_id = sc.id
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
    id UUID, name VARCHAR(64), email VARCHAR(320), type VARCHAR(128), verified BOOL, avatar VARCHAR(128), country VARCHAR(2), company VARCHAR(256), jobTitle VARCHAR(128)
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, coalesce(u.email, ''), u.type, u.verified, u.avatar, u.country, u.company, u.job_title FROM users u WHERE u.id = userId;
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
			w.id, w.name, su.active
		FROM storyboard_user su
		LEFT JOIN users w ON su.user_id = w.id
		WHERE su.storyboard_id = storyboardId
		ORDER BY w.name;
END;
$$ LANGUAGE plpgsql;

-- Get Storyboard Personas
DROP FUNCTION IF EXISTS get_storyboard_personas(uuid);
CREATE FUNCTION get_storyboard_personas(storyboardId UUID) RETURNS table (
    id UUID,
    name VARCHAR(256),
    role VARCHAR(256),
    description TEXT
) AS $$
BEGIN
    RETURN QUERY
        SELECT
			p.id, p.name, p.role, p.description
		FROM storyboard_persona p
		WHERE p.storyboard_id = storyboardId;
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
			w.id, w.name, coalesce(su.active, FALSE)
		FROM users w
		LEFT JOIN storyboard_user su ON su.user_id = w.id AND su.storyboard_id = storyboardId
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
DROP FUNCTION IF EXISTS get_app_stats(
    OUT unregistered_user_count INTEGER,
    OUT registered_user_count INTEGER,
    OUT storyboard_count INTEGER
);
DROP FUNCTION IF EXISTS get_app_stats(
    OUT unregistered_user_count INTEGER,
    OUT registered_user_count INTEGER,
    OUT storyboard_count INTEGER,
    OUT organization_count INTEGER,
    OUT department_count INTEGER,
    OUT team_count INTEGER
);
CREATE FUNCTION get_app_stats(
    OUT unregistered_user_count INTEGER,
    OUT registered_user_count INTEGER,
    OUT storyboard_count INTEGER,
    OUT organization_count INTEGER,
    OUT department_count INTEGER,
    OUT team_count INTEGER,
    OUT apikey_count INTEGER
) AS $$
BEGIN
    SELECT COUNT(*) INTO unregistered_user_count FROM users WHERE email IS NULL;
    SELECT COUNT(*) INTO registered_user_count FROM users WHERE email IS NOT NULL;
    SELECT COUNT(*) INTO storyboard_count FROM storyboard;
    SELECT COUNT(*) INTO organization_count FROM organization;
    SELECT COUNT(*) INTO department_count FROM organization_department;
    SELECT COUNT(*) INTO team_count FROM team;
    SELECT COUNT(*) INTO apikey_count FROM api_keys;
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
    IF FOUND THEN
        INSERT INTO user_reset (user_id) VALUES (userId) RETURNING reset_id INTO resetId;
    ELSE
        RAISE EXCEPTION 'Nonexistent User --> %', userEmail USING HINT = 'Please check your Email';
    END IF;
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
         last_active = NOW(),
         updated_date = NOW()
    WHERE id = activeUserId
    RETURNING id INTO userId;

    INSERT INTO user_verify (user_id) VALUES (userId) RETURNING verify_id INTO verifyId;
END;
$$ LANGUAGE plpgsql;

-- Get a list of countries
CREATE OR REPLACE FUNCTION countries_active() RETURNS table (
    country VARCHAR(2)
) AS $$
BEGIN
    RETURN QUERY SELECT ac.country FROM active_countries ac;
END;
$$ LANGUAGE plpgsql;

--
-- ORGANIZATIONS --
--

-- Get Organization --
CREATE OR REPLACE FUNCTION organization_get_by_id(
    IN orgId UUID
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT o.id, o.name, o.created_date, o.updated_date
        FROM organization o
        WHERE o.id = orgId;
END;
$$ LANGUAGE plpgsql;

-- Get Organization User Role --
CREATE OR REPLACE FUNCTION organization_get_user_role(
    IN userId UUID,
    IN orgId UUID,
    OUT role VARCHAR(16)
) AS $$
BEGIN
    SELECT ou.role INTO role
    FROM organization_user ou
    WHERE ou.organization_id = orgId AND ou.user_id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get Organizations --
CREATE OR REPLACE FUNCTION organization_list(
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table(
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT o.id, o.name, o.created_date, o.updated_date
        FROM organization o
        ORDER BY o.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Get Organizations by User --
CREATE OR REPLACE FUNCTION organization_list_by_user(
    IN userId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP, role VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT o.id, o.name, o.created_date, o.updated_date, ou.role
        FROM organization_user ou
        LEFT JOIN organization o ON ou.organization_id = o.id
        WHERE ou.user_id = userId
        ORDER BY created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Create Organization --
CREATE OR REPLACE FUNCTION organization_create(
    IN userId UUID,
    IN orgName VARCHAR(256),
    OUT organizationId UUID
) AS $$
BEGIN
    INSERT INTO organization (name) VALUES (orgName) RETURNING id INTO organizationId;
    INSERT INTO organization_user (organization_id, user_id, role) VALUES (organizationId, userId, 'ADMIN');
END;
$$ LANGUAGE plpgsql;

-- Add User to Organization --
CREATE OR REPLACE FUNCTION organization_user_add(
    IN orgId UUID,
    IN userId UUID,
    IN userRole VARCHAR(16)
) RETURNS void AS $$
BEGIN
    INSERT INTO organization_user (organization_id, user_id, role) VALUES (orgId, userId, userRole);
    UPDATE organization SET updated_date = NOW() WHERE id = orgId;
END;
$$ LANGUAGE plpgsql;

-- Remove User from Organization --
CREATE OR REPLACE PROCEDURE organization_user_remove(orgId UUID, userId UUID)
AS $$
DECLARE temprow record;
BEGIN
    FOR temprow IN
        SELECT id FROM organization_department WHERE organization_id = orgId
    LOOP
        CALL department_user_remove(temprow.id, userId);
    END LOOP;
    DELETE FROM team_user tu WHERE tu.team_id IN (
        SELECT ot.team_id
        FROM organization_team ot
        WHERE ot.organization_id = orgId
    ) AND tu.user_id = userId;
    DELETE FROM organization_user WHERE organization_id = orgId AND user_id = userId;
    UPDATE organization SET updated_date = NOW() WHERE id = orgId;

    COMMIT;
END;
$$ LANGUAGE plpgsql;

-- Get Organization Users --
CREATE OR REPLACE FUNCTION organization_user_list(
    IN orgId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), email VARCHAR(256), role VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, u.email, ou.role
        FROM organization_user ou
        LEFT JOIN users u ON ou.user_id = u.id
        WHERE ou.organization_id = orgId
        ORDER BY ou.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Get Organization Teams --
CREATE OR REPLACE FUNCTION organization_team_list(
    IN orgId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.name, t.created_date, t.updated_date
        FROM organization_team ot
        LEFT JOIN team t ON ot.team_id = t.id
        WHERE ot.organization_id = orgId
        ORDER BY t.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Create Organization Team --
CREATE OR REPLACE FUNCTION organization_team_create(
    IN orgId UUID,
    IN teamName VARCHAR(256),
    OUT teamId UUID
) AS $$
BEGIN
    INSERT INTO team (name) VALUES (teamName) RETURNING id INTO teamId;
    INSERT INTO organization_team (organization_id, team_id) VALUES (orgId, teamId);
    UPDATE organization SET updated_date = NOW() WHERE id = orgId;
END;
$$ LANGUAGE plpgsql;

-- Get Organization Team User Role --
CREATE OR REPLACE FUNCTION organization_team_user_role(
    IN userId UUID,
    IN orgId UUID,
    IN teamId UUID
) RETURNS table (
    orgRole VARCHAR(16), teamRole VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT ou.role AS orgRole, COALESCE(tu.role, '') AS teamRole
        FROM organization_user ou
        LEFT JOIN team_user tu ON tu.user_id = userId AND tu.team_id = teamId
        WHERE ou.organization_id = orgId AND ou.user_id = userId;
END;
$$ LANGUAGE plpgsql;

--
-- DEPARTMENTS --
--

-- Get Department --
CREATE OR REPLACE FUNCTION department_get_by_id(
    IN departmentId UUID
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT od.id, od.name, od.created_date, od.updated_date
        FROM organization_department od
        WHERE od.id = departmentId;
END;
$$ LANGUAGE plpgsql;

-- Get Department User Role --
CREATE OR REPLACE FUNCTION department_get_user_role(
    IN userId UUID,
    IN orgId UUID,
    IN departmentId UUID
) RETURNS table (
    orgRole VARCHAR(16), departmentRole VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT ou.role AS orgRole, COALESCE(du.role, '') AS departmentRole
        FROM organization_user ou
        LEFT JOIN department_user du ON du.user_id = userId AND du.department_id = departmentId
        WHERE ou.organization_id = orgId AND ou.user_id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get Organization Departments --
CREATE OR REPLACE FUNCTION department_list(
    IN orgId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT d.id, d.name, d.created_date, d.updated_date
        FROM organization_department d
        WHERE d.organization_id = orgId
        ORDER BY d.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Create Organization Department --
CREATE OR REPLACE FUNCTION department_create(
    IN orgId UUID,
    IN departmentName VARCHAR(256),
    OUT departmentId UUID
) AS $$
BEGIN
    INSERT INTO organization_department (name, organization_id) VALUES (departmentName, orgId) RETURNING id INTO departmentId;
    UPDATE organization SET updated_date = NOW() WHERE id = orgId;
END;
$$ LANGUAGE plpgsql;

-- Get Department Teams --
CREATE OR REPLACE FUNCTION department_team_list(
    IN departmentId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.name, t.created_date, t.updated_date
        FROM department_team dt
        LEFT JOIN team t ON dt.team_id = t.id
        WHERE dt.department_id = departmentId
        ORDER BY t.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Create Department Team --
CREATE OR REPLACE FUNCTION department_team_create(
    IN departmentId UUID,
    IN teamName VARCHAR(256),
    OUT teamId UUID
) AS $$
BEGIN
    INSERT INTO team (name) VALUES (teamName) RETURNING id INTO teamId;
    INSERT INTO department_team (department_id, team_id) VALUES (departmentId, teamId);
    UPDATE organization_department SET updated_date = NOW() WHERE id = departmentId;
END;
$$ LANGUAGE plpgsql;

-- Get Department Team User Role --
CREATE OR REPLACE FUNCTION department_team_user_role(
    IN userId UUID,
    IN orgId UUID,
    IN departmentId UUID,
    IN teamId UUID
) RETURNS table (
    orgRole VARCHAR(16), departmentRole VARCHAR(16), teamRole VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT ou.role AS orgRole, COALESCE(du.role, '') AS departmentRole, COALESCE(tu.role, '') AS teamRole
        FROM organization_user ou
        LEFT JOIN department_user du ON du.user_id = userId AND du.department_id = departmentId
        LEFT JOIN team_user tu ON tu.user_id = userId AND tu.team_id = teamId
        WHERE ou.organization_id = orgId AND ou.user_id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get Department Users --
CREATE OR REPLACE FUNCTION department_user_list(
    IN departmentId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), email VARCHAR(256), role VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, u.email, du.role
        FROM department_user du
        LEFT JOIN users u ON du.user_id = u.id
        WHERE du.department_id = departmentId
        ORDER BY du.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Add User to Department --
CREATE OR REPLACE FUNCTION department_user_add(
    IN departmentId UUID,
    IN userId UUID,
    IN userRole VARCHAR(16)
) RETURNS void AS $$
DECLARE orgId UUID;
BEGIN    
    SELECT organization_id INTO orgId FROM organization_user WHERE user_id = userId;

    IF orgId IS NULL THEN
        RAISE EXCEPTION 'User not in Organization -> %', userId USING HINT = 'Please add user to Organization before department';
    END IF;

    INSERT INTO department_user (department_id, user_id, role) VALUES (departmentId, userId, userRole);
    UPDATE organization_department SET updated_date = NOW() WHERE id = departmentId;
END;
$$ LANGUAGE plpgsql;

-- Remove User from Department --
CREATE OR REPLACE PROCEDURE department_user_remove(departmentId UUID, userId UUID)
AS $$
BEGIN
    DELETE FROM team_user tu WHERE tu.team_id IN (
        SELECT dt.team_id
        FROM department_team dt
        WHERE dt.department_id = departmentId
    ) AND tu.user_id = userId;
    DELETE FROM department_user WHERE department_id = departmentId AND user_id = userId;
    UPDATE organization_department SET updated_date = NOW() WHERE id = departmentId;

    COMMIT;
END;
$$ LANGUAGE plpgsql;

--
-- TEAMS --
--

-- Get Team --
CREATE OR REPLACE FUNCTION team_get_by_id(
    IN teamId UUID
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT o.id, o.name, o.created_date, o.updated_date
        FROM team o
        WHERE o.id = teamId;
END;
$$ LANGUAGE plpgsql;

-- Get Team User Role --
CREATE OR REPLACE FUNCTION team_get_user_role(
    IN userId UUID,
    IN teamId UUID
) RETURNS table (
    role VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT tu.role
        FROM team_user tu
        WHERE tu.team_id = teamId AND tu.user_id = userId;
END;
$$ LANGUAGE plpgsql;

-- Get Teams --
CREATE OR REPLACE FUNCTION team_list(
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.name, t.created_date, t.updated_date
        FROM team t
        ORDER BY t.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Get Teams by User --
CREATE OR REPLACE FUNCTION team_list_by_user(
    IN userId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), created_date TIMESTAMP, updated_date TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.name, t.created_date, t.updated_date
        FROM team_user tu
        LEFT JOIN team t ON tu.team_id = t.id
        WHERE tu.user_id = userId
        ORDER BY t.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Create Team --
CREATE OR REPLACE FUNCTION team_create(
    IN userId UUID,
    IN teamName VARCHAR(256),
    OUT teamId UUID
) AS $$
BEGIN
    INSERT INTO team (name) VALUES (teamName) RETURNING id INTO teamId;
    INSERT INTO team_user (team_id, user_id, role) VALUES (teamId, userId, 'ADMIN');
END;
$$ LANGUAGE plpgsql;

-- Get Team Users --
CREATE OR REPLACE FUNCTION team_user_list(
    IN teamId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256), email VARCHAR(256), role VARCHAR(16)
) AS $$
BEGIN
    RETURN QUERY
        SELECT u.id, u.name, u.email, tu.role
        FROM team_user tu
        LEFT JOIN users u ON tu.user_id = u.id
        WHERE tu.team_id = teamId
        ORDER BY tu.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Add User to Team --
CREATE OR REPLACE FUNCTION team_user_add(
    IN teamId UUID,
    IN userId UUID,
    IN userRole VARCHAR(16)
) RETURNS void AS $$
BEGIN
    INSERT INTO team_user (team_id, user_id, role) VALUES (teamId, userId, userRole);
    UPDATE team SET updated_date = NOW() WHERE id = teamId;
END;
$$ LANGUAGE plpgsql;

-- Remove User from Team --
CREATE OR REPLACE PROCEDURE team_user_remove(teamId UUID, userId UUID)
AS $$
BEGIN
    DELETE FROM team_user WHERE team_id = teamId AND user_id = userId;
    UPDATE team SET updated_date = NOW() WHERE id = teamId;
END;
$$ LANGUAGE plpgsql;

-- Get Team Storyboards --
CREATE OR REPLACE FUNCTION team_storyboard_list(
    IN teamId UUID,
    IN l_limit INTEGER,
    IN l_offset INTEGER
) RETURNS table (
    id UUID, name VARCHAR(256)
) AS $$
BEGIN
    RETURN QUERY
        SELECT b.id, b.name
        FROM team_storyboard tb
        LEFT JOIN storyboard b ON tb.storyboard_id = b.id
        WHERE tb.team_id = teamId
        ORDER BY tb.created_date
		LIMIT l_limit
		OFFSET l_offset;
END;
$$ LANGUAGE plpgsql;

-- Add Storyboard to Team --
CREATE OR REPLACE FUNCTION team_storyboard_add(
    IN teamId UUID,
    IN storyboardId UUID
) RETURNS void AS $$
BEGIN
    INSERT INTO team_storyboard (team_id, storyboard_id) VALUES (teamId, storyboardId);
    UPDATE team SET updated_date = NOW() WHERE id = teamId;
END;
$$ LANGUAGE plpgsql;

-- Remove Storyboard from Team --
CREATE OR REPLACE FUNCTION team_storyboard_remove(
    IN teamId UUID,
    IN storyboardId UUID
) RETURNS void AS $$
BEGIN
    DELETE FROM team_storyboard WHERE storyboard_id = storyboardId AND team_id = teamId;
    UPDATE team SET updated_date = NOW() WHERE id = teamId;
END;
$$ LANGUAGE plpgsql;

-- Delete Team --
CREATE OR REPLACE PROCEDURE team_delete(teamId UUID)
AS $$
BEGIN
    DELETE FROM team WHERE id = teamId;
END;
$$ LANGUAGE plpgsql;