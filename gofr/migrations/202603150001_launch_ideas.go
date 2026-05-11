package migrations

import "gofr.dev/pkg/gofr/migration"

const launchIdeasFeature = `
CREATE TABLE IF NOT EXISTS launch_ideas (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	title TEXT NOT NULL,
	pitch TEXT NOT NULL,
	stage TEXT NOT NULL,
	hype_score INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_boosted_at TIMESTAMPTZ
);

INSERT INTO launch_ideas (title, pitch, stage, hype_score, last_boosted_at)
VALUES
	('AI release notes from commits', 'Turns merged pull requests into human-readable release notes.', 'prototype', 3, NOW() - INTERVAL '2 hours'),
	('Stand-up summarizer for Slack', 'Builds a daily stand-up digest from team updates and blockers.', 'beta', 5, NOW() - INTERVAL '30 minutes'),
	('Incident timeline generator', 'Pulls deploys, alerts and commits into one incident timeline.', 'discovery', 1, NULL)
ON CONFLICT DO NOTHING;
`

func addLaunchIdeasFeature() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			_, err := d.SQL.Exec(launchIdeasFeature)
			return err
		},
	}
}
