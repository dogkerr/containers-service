CREATE TYPE container_status AS ENUM ('RUN', 'STOP');


CREATE TABLE users (
    id  UUID DEFAULT gen_random_uuid() PRIMARY KEY ,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL, 
    fullname VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE TABLE containers (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY ,
    user_id UUID NOT NULL,
    image VARCHAR(255) NOT NULL,
    status container_status NOT NULL,
    name VARCHAR(255) NOT NULL,
    container_port int NOT NULL,
    public_port int,
    terminated_time timestamp with time zone ,
    created_time timestamp with time zone  DEFAULT NOW() NOT NULL,
    service_id VARCHAR(255) NOT NULL
);




CREATE TABLE container_lifecycles (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY ,
    container_id UUID,
    start_time TIMESTAMP with time zone NOT NULL,
    stop_time TIMESTAMP with time zone,
    status container_status NOT NULL,
    replica INT NOT NULL
);

CREATE TABLE container_metrics (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY ,
    container_id UUID NOT NULL,
    cpus FLOAT NOT NULL,
    memory FLOAT NOT NULL,
    network_ingress FlOAT NOT NULL,
    network_egress FLOAT NOT NULL
);


ALTER TABLE container_lifecycles ADD  CONSTRAINT fk_lifecycles_containers
    FOREIGN KEY (container_id)
    REFERENCES containers (id);

ALTER TABLE containers ADD CONSTRAINT fk_containers_users
    FOREIGN KEY (user_id) 
    REFERENCES users(id);


ALTER TABLE container_metrics ADD CONSTRAINT fk_container_metrics 
    FOREIGN KEY (container_id)
    REFERENCES containers(id);


INSERT INTO users( username, email, password) 
    VALUES ('asda', 'sadasd@gmail.com', 'asdad');

CREATE TYPE dashboard_type AS ENUM('log', 'monitor');

CREATE TABLE dashboards (
    id  UUID DEFAULT gen_random_uuid() PRIMARY KEY ,
    uid VARCHAR(255) NOT NULL,
    owner UUID NOT NULL,
    db_type dashboard_type NOT NULL
);

ALTER TABLE dashboards ADD CONSTRAINT fk_dashboards_users
    FOREIGN KEY (owner)
    REFERENCES users(id);

    
