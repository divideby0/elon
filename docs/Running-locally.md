_Note: this doc is in progress_

To run locally, you need a local MySQL and a local Sysbreaker. This page
describes how to start both of those up using Docker containers

## MySQL

This will start up a MySQL container with the root password as `password`.

```bash
docker run -e MYSQL_ROOT_PASSWORD=password -p3306:3306 mysql:5.6
```
