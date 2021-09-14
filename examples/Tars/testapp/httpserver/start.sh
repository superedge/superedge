#设置set信息
deploy=$DEPLOY_NAME
str=$POD_NAME
str=${str#${deploy}-*}
str=${str%-*}
str=${str%-*}
tars_app=`tr '[A-Z]' '[a-z]' <<<"$TARS_APP"`
export TARS_SET_ID=$TARS_APP.${str}.1

#启动服务
tarscli supervisor
