FROM mysql:8.0.28

COPY ./docker_env/mysql/my.cnf /etc/mysql/conf.d/my.cnf
