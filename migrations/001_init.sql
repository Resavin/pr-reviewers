CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE teams (
    team_name text PRIMARY KEY
);

CREATE TABLE users (
    user_id   text PRIMARY KEY,
    username  text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    team_name text NOT NULL REFERENCES teams(team_name)
);

CREATE TABLE pull_requests (
    pull_request_id   text PRIMARY KEY,
    pull_request_name text NOT NULL,
    author_id         text NOT NULL REFERENCES users(user_id),
    status            pr_status NOT NULL DEFAULT 'OPEN',
    created_at        timestamptz NOT NULL DEFAULT now(),
    merged_at         timestamptz
);

-- (0..2 for PR)
CREATE TABLE pull_request_reviewers (
    pull_request_id text NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     text NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

-- make /users/getReview faster
CREATE INDEX idx_pr_reviewers_reviewer ON pull_request_reviewers(reviewer_id);
