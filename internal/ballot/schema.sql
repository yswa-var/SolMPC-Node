-- Ballot Management System Database Schema
-- This schema implements the database structure for the ballot management system
-- as specified in Phase 2 of the production roadmap

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ballots table - stores ballot information and metadata
CREATE TABLE ballots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'multiple_choice', -- yes_no, multiple_choice, ranked, custom
    options JSONB NOT NULL, -- Array of ballot options with id, text, description, order
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    eligible_voters JSONB, -- Array of wallet addresses or voter IDs
    status VARCHAR(20) DEFAULT 'draft', -- draft, active, closed, archived
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL, -- Creator's wallet address
    requires_auth BOOLEAN DEFAULT true,
    allow_anonymous BOOLEAN DEFAULT false,
    max_votes_per_voter INTEGER DEFAULT 1,
    metadata JSONB, -- Additional custom data
    
    -- Constraints
    CONSTRAINT valid_status CHECK (status IN ('draft', 'active', 'closed', 'archived')),
    CONSTRAINT valid_type CHECK (type IN ('yes_no', 'multiple_choice', 'ranked', 'custom')),
    CONSTRAINT valid_timing CHECK (end_time > start_time),
    CONSTRAINT valid_max_votes CHECK (max_votes_per_voter >= 1)
);

-- Votes table - stores individual votes cast by voters
CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ballot_id UUID NOT NULL REFERENCES ballots(id) ON DELETE CASCADE,
    voter_id VARCHAR(255) NOT NULL, -- Wallet address or anonymous ID
    choices JSONB NOT NULL, -- Array of choices with option_id, rank, weight
    timestamp TIMESTAMP DEFAULT NOW(),
    signature TEXT, -- Cryptographic signature for verification
    is_anonymous BOOLEAN DEFAULT false,
    
    -- Constraints
    UNIQUE(ballot_id, voter_id) -- Prevent duplicate votes from same voter on same ballot
);

-- Tally results table - stores final voting results
CREATE TABLE tally_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ballot_id UUID NOT NULL UNIQUE REFERENCES ballots(id) ON DELETE CASCADE,
    total_votes INTEGER NOT NULL DEFAULT 0,
    results JSONB NOT NULL, -- Map of option_id to vote count
    rankings JSONB, -- For ranked choice voting results
    computed_at TIMESTAMP DEFAULT NOW(),
    mpc_signature TEXT, -- MPC signature for result verification
    proof_hash TEXT, -- Hash of cryptographic proof
    metadata JSONB -- Additional result metadata
);

-- Ballot templates table - stores reusable ballot configurations
CREATE TABLE ballot_templates (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    options JSONB NOT NULL,
    settings JSONB NOT NULL, -- Duration, auth requirements, etc.
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_template_type CHECK (type IN ('yes_no', 'multiple_choice', 'ranked', 'custom'))
);

-- Voter eligibility table - manages voter access control
CREATE TABLE voter_eligibility (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ballot_id UUID NOT NULL REFERENCES ballots(id) ON DELETE CASCADE,
    voter_id VARCHAR(255) NOT NULL,
    is_whitelisted BOOLEAN DEFAULT true,
    is_blacklisted BOOLEAN DEFAULT false,
    min_balance BIGINT, -- Minimum token balance required
    token_address VARCHAR(255), -- Token contract for balance check
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(ballot_id, voter_id)
);

-- Audit log table - tracks all voting actions for transparency
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ballot_id UUID REFERENCES ballots(id) ON DELETE SET NULL,
    vote_id UUID REFERENCES votes(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- create_ballot, cast_vote, close_ballot, etc.
    actor VARCHAR(255) NOT NULL, -- Wallet address of actor
    details JSONB, -- Additional action details
    timestamp TIMESTAMP DEFAULT NOW(),
    
    -- Index for efficient querying
    INDEX idx_audit_ballot_id ON audit_logs(ballot_id),
    INDEX idx_audit_timestamp ON audit_logs(timestamp)
);

-- Validator sessions table - tracks MPC validator participation
CREATE TABLE validator_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_type VARCHAR(50) NOT NULL, -- keygen, signing, tally
    ballot_id UUID REFERENCES ballots(id) ON DELETE SET NULL,
    validator_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, completed, failed
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    mpc_data JSONB, -- MPC-specific session data
    
    CONSTRAINT valid_session_type CHECK (session_type IN ('keygen', 'signing', 'tally')),
    CONSTRAINT valid_session_status CHECK (status IN ('active', 'completed', 'failed'))
);

-- MPC sessions table - tracks multi-party computation sessions
CREATE TABLE mpc_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) NOT NULL UNIQUE,
    operation_type VARCHAR(50) NOT NULL, -- key_generation, threshold_signing, vote_tally
    ballot_id UUID REFERENCES ballots(id) ON DELETE SET NULL,
    participants JSONB NOT NULL, -- Array of participant validator IDs
    threshold INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'initializing',
    result_data JSONB, -- Final MPC computation result
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    CONSTRAINT valid_mpc_operation CHECK (operation_type IN ('key_generation', 'threshold_signing', 'vote_tally')),
    CONSTRAINT valid_mpc_status CHECK (status IN ('initializing', 'running', 'completed', 'failed'))
);

-- Indexes for performance optimization
CREATE INDEX idx_ballots_status ON ballots(status);
CREATE INDEX idx_ballots_created_by ON ballots(created_by);
CREATE INDEX idx_ballots_timing ON ballots(start_time, end_time);
CREATE INDEX idx_votes_ballot_id ON votes(ballot_id);
CREATE INDEX idx_votes_voter_id ON votes(voter_id);
CREATE INDEX idx_votes_timestamp ON votes(timestamp);
CREATE INDEX idx_tally_results_ballot_id ON tally_results(ballot_id);
CREATE INDEX idx_validator_sessions_ballot_id ON validator_sessions(ballot_id);
CREATE INDEX idx_mpc_sessions_ballot_id ON mpc_sessions(ballot_id);

-- Triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ballots_updated_at 
    BEFORE UPDATE ON ballots 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically create audit log entries
CREATE OR REPLACE FUNCTION create_audit_log()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (ballot_id, vote_id, action, actor, details)
        VALUES (
            COALESCE(NEW.ballot_id, NEW.id),
            CASE WHEN TG_TABLE_NAME = 'votes' THEN NEW.id ELSE NULL END,
            TG_OP || '_' || TG_TABLE_NAME,
            COALESCE(NEW.created_by, NEW.voter_id, 'system'),
            row_to_json(NEW)
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (ballot_id, vote_id, action, actor, details)
        VALUES (
            COALESCE(NEW.ballot_id, NEW.id),
            CASE WHEN TG_TABLE_NAME = 'votes' THEN NEW.id ELSE NULL END,
            TG_OP || '_' || TG_TABLE_NAME,
            COALESCE(NEW.created_by, NEW.voter_id, 'system'),
            jsonb_build_object('old', row_to_json(OLD), 'new', row_to_json(NEW))
        );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (ballot_id, vote_id, action, actor, details)
        VALUES (
            COALESCE(OLD.ballot_id, OLD.id),
            CASE WHEN TG_TABLE_NAME = 'votes' THEN OLD.id ELSE NULL END,
            TG_OP || '_' || TG_TABLE_NAME,
            'system',
            row_to_json(OLD)
        );
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit triggers for key tables
CREATE TRIGGER audit_ballots 
    AFTER INSERT OR UPDATE OR DELETE ON ballots 
    FOR EACH ROW 
    EXECUTE FUNCTION create_audit_log();

CREATE TRIGGER audit_votes 
    AFTER INSERT OR UPDATE OR DELETE ON votes 
    FOR EACH ROW 
    EXECUTE FUNCTION create_audit_log();

-- Insert default ballot templates
INSERT INTO ballot_templates (id, name, description, type, options, settings) VALUES
('yes_no', 'Yes/No Vote', 'Simple yes or no question', 'yes_no', 
 '[{"id":"yes","text":"Yes","description":"Vote in favor","order":1},{"id":"no","text":"No","description":"Vote against","order":2}]',
 '{"duration":"24:00:00","requires_auth":true,"allow_anonymous":false,"max_votes_per_voter":1}'),

('multiple_choice', 'Multiple Choice', 'Choose one option from multiple choices', 'multiple_choice',
 '[{"id":"option_a","text":"Option A","order":1},{"id":"option_b","text":"Option B","order":2},{"id":"option_c","text":"Option C","order":3}]',
 '{"duration":"48:00:00","requires_auth":true,"allow_anonymous":false,"max_votes_per_voter":1}'),

('ranked_choice', 'Ranked Choice Voting', 'Rank options in order of preference', 'ranked',
 '[{"id":"candidate_1","text":"Candidate 1","order":1},{"id":"candidate_2","text":"Candidate 2","order":2},{"id":"candidate_3","text":"Candidate 3","order":3}]',
 '{"duration":"72:00:00","requires_auth":true,"allow_anonymous":false,"max_votes_per_voter":3}'),

('approval_voting', 'Approval Voting', 'Select all options you approve of', 'multiple_choice',
 '[{"id":"proposal_a","text":"Proposal A","order":1},{"id":"proposal_b","text":"Proposal B","order":2},{"id":"proposal_c","text":"Proposal C","order":3},{"id":"proposal_d","text":"Proposal D","order":4}]',
 '{"duration":"48:00:00","requires_auth":true,"allow_anonymous":false,"max_votes_per_voter":4}'),

('anonymous_poll', 'Anonymous Poll', 'Anonymous voting with privacy protection', 'yes_no',
 '[{"id":"yes","text":"Yes","order":1},{"id":"no","text":"No","order":2}]',
 '{"duration":"24:00:00","requires_auth":false,"allow_anonymous":true,"max_votes_per_voter":1}');

-- Create views for common queries
CREATE VIEW active_ballots AS
SELECT b.*, 
       (SELECT COUNT(*) FROM votes v WHERE v.ballot_id = b.id) as vote_count
FROM ballots b 
WHERE b.status = 'active' 
  AND b.start_time <= NOW() 
  AND b.end_time > NOW();

CREATE VIEW ballot_results AS
SELECT b.id, b.title, b.status, 
       tr.total_votes, tr.results, tr.computed_at, tr.mpc_signature
FROM ballots b
LEFT JOIN tally_results tr ON b.id = tr.ballot_id
WHERE b.status IN ('closed', 'archived');

-- Grant permissions (adjust as needed for your user)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ballot_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ballot_user;