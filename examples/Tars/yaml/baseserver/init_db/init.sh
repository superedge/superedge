#!/bin/bash

export pod=<改为mysql pod的名字> 
export user=root
export pass=pass
export host=localhost

kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database db_tars" || exit 1

kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database db_user_system"
kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database db_tars_web"
kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database db_cache_web"
kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database tars_stat"
kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} -e "create database tars_property"

kubectl exec -i ${pod} -- mysql -h ${host} -u${user} -p${pass} db_tars < ./db_tars.sql
kubectl exec -i ${pod} -- mysql -h ${host} -u${user} -p${pass} db_user_system < ./db_user_system.sql
kubectl exec -i ${pod} -- mysql -h ${host} -u${user} -p${pass} db_tars_web < ./db_tars_web.sql
kubectl exec -i ${pod} -- mysql -h ${host} -u${user} -p${pass} db_cache_web < ./db_cache_web.sql
kubectl exec -i ${pod} -- mysql -h ${host} -u${user} -p${pass} tars_stat < ./db_cache_web.sql

kubectl exec ${pod} -- mysql -h ${host} -u${user} -p${pass} db_tars -e "show tables"
