DROP TABLE IF EXISTS request_response;

CREATE TABLE IF NOT EXISTS request_response
(
    request_response_id SERIAL NOT NULL PRIMARY KEY,
    request    TEXT   NOT NULL,
    response    TEXT  NOT NULL
);
