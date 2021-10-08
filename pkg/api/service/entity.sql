

'kcore_entity table for mysql.'
CREATE TABLE IF NOT EXISTS kcore_entity(
   id varchar(127) UNIQUE NOT NULL,
   user_id VARCHAR(63) NOT NULL,
   source VARCHAR(63) NOT NULL,
   tag VARCHAR(63),
   status VARCHAR(63),
   version INTEGER,
   entity_key VARCHAR(127),
   deleted_id VARCHAR(255),
   PRIMARY KEY ( id )
)ENGINE=InnoDB DEFAULT CHARSET=utf8;