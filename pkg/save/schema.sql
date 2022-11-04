CREATE TABLE person
( handle TEXT PRIMARY KEY
);
CREATE TABLE interval
( person TEXT NOT NULL
, start_at TIMESTAMP NOT NULL
, end_before TIMESTAMP NOT NULL
, kind TEXT NOT NULL
, FOREIGN KEY (person) REFERENCES person(handle)
, CHECK ( start_at < end_before )
, CHECK ( kind IN
		('SHIFT'
		,'EXCLUSION'
		)
	)
);
