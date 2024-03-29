-- Drop existing tables if they exist
DROP TABLE IF EXISTS ASSET;
DROP TABLE IF EXISTS TASK;
DROP TABLE IF EXISTS WORKFLOW;
DROP TABLE IF EXISTS PROJECT;

-- Project Table
CREATE TABLE PROJECT (
                         ID VARCHAR(255) PRIMARY KEY,
                         NAME VARCHAR(255) NOT NULL,
                         DESCRIPTION TEXT
) ENGINE=InnoDB;

-- Workflow Table
CREATE TABLE WORKFLOW (
                          ID VARCHAR(255) PRIMARY KEY,
                          PARENT_ID VARCHAR(255),
                          POSITION INT,
                          REVISION VARCHAR(255),
                          URI TEXT NOT NULL ,
                          PROJECT_ID VARCHAR(255) NOT NULL,
                          NAME VARCHAR(255) NOT NULL ,
                          INIT TEXT , -- JSON Array
                          POST TEXT , -- JSON Array
                          TEMPLATE VARCHAR(255),
                          INSTANCE_INDEX INT,
                          INSTANCE_TAG VARCHAR(255),
                          FOREIGN KEY(PROJECT_ID) REFERENCES PROJECT(ID)
) ENGINE=InnoDB;

-- Task Table
CREATE TABLE TASK (
                      ID VARCHAR(255) PRIMARY KEY,
                      WORKFLOW_ID VARCHAR(255) NOT NULL,
                      PARENT_ID VARCHAR(255),
                      POSITION INT,
                      TAG VARCHAR(255) NOT NULL,
                      INIT TEXT  , -- JSON Array
                      POST TEXT,   -- JSON Array
                      DESCRIPTION TEXT,
                      WHEN_EXPR TEXT  ,
                      EXIT_EXPR TEXT  ,
                      ON_ERROR TEXT  ,
                      DEFERRED TEXT  ,
                      SERVICE VARCHAR(255),
                      ACTION VARCHAR(255),
                      INPUT TEXT,
                      INPUT_URI TEXT,
                      ASYNC BOOLEAN,
                      SKIP_EXPR TEXT,
                      IS_TEMPLATE BOOLEAN,
                      SUB_PATH VARCHAR(255),
                      RANGE_EXPR TEXT,
                      DATA TEXT, -- JSON Object
                      VARIABLES TEXT, -- JSON Array
                      EXTRACTS TEXT, -- JSON Object
                      SLEEP_TIME_MS INT,
                      THINK_TIME_MS INT,
                      LOGGING BOOLEAN,
                      REPEAT_RUN INT,
                      INSTANCE_INDEX INT,
                      INSTANCE_TAG VARCHAR(255)
) ENGINE=InnoDB;

-- Asset Table
CREATE TABLE ASSET (
                       ID VARCHAR(255) PRIMARY KEY,
                       LOCATION VARCHAR(255) NOT NULL,
                       DESCRIPTION TEXT,
                       WORKFLOW_ID VARCHAR(255) NOT NULL,
                       IS_DIR BOOLEAN,
                       TEMPLATE VARCHAR(255),
                       INSTANCE_INDEX INT,
                       INSTANCE_TAG VARCHAR(255),
                       POSITION INT,
                       SOURCE BLOB,
                       FORMAT VARCHAR(255) ,
                       CODEC VARCHAR(255),
                       FOREIGN KEY(WORKFLOW_ID) REFERENCES WORKFLOW(ID)
) ENGINE=InnoDB;
