#!/bin/sh
MYSQL_PWD=1 mysql -h localhost -P 3306 --protocol=tcp -u root todoapp < schema.sql 
