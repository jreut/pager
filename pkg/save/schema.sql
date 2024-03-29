CREATE TABLE schedule
( name TEXT PRIMARY KEY
, CHECK ( name != '' )
);
CREATE TABLE event
( person TEXT NOT NULL
, schedule TEXT NOT NULL
, kind TEXT NOT NULL
, at TIMESTAMP NOT NULL
, FOREIGN KEY (schedule) REFERENCES schedule(name)
, CHECK ( person != '' )
, CHECK ( kind IN
		('ADD'
		,'REMOVE'
		)
	)
);
CREATE TABLE interval
( person TEXT NOT NULL
, schedule TEXT NOT NULL
, start_at TIMESTAMP NOT NULL
, end_before TIMESTAMP NOT NULL
, kind TEXT NOT NULL
, FOREIGN KEY (schedule) REFERENCES schedule(name)
, CHECK ( person != '' )
, CHECK ( start_at < end_before )
, CHECK ( kind IN
		('SHIFT'
		,'EXCLUSION'
		)
	)
);
