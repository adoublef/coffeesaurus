-- https://www.sqlite.org/lang_datefunc.html
-- https://sqlite.org/forum/forumpost/f3c7cac8a7
-- https://lukas-r.blog/posts/2023-06-03-searching-for-user-sessions/
-- https://topic.alibabacloud.com/a/the-time-difference-judgment-of-sqlite-two-ways-to-delete-data-from-n-days-ago_1_43_30031034.html
CREATE TABLE sessions (
    -- profile is uuid formatted
    `id` TEXT,
    -- profile is xid formatted
    `profile` TEXT NOT NULL,
    -- TODO `created_at`
    -- TODO `expire_at`
    --
    PRIMARY KEY (id, profile)
);