# lgtm-gitlab

lgtm-gitlab is used to auto merge gitlab CE MR with your LGTM like [gitlab EE approve](https://about.gitlab.com/2015/06/16/feature-highlight-approve-merge-request/)

# usage

## access token

You should create a access token on your gitlab

## run lgtm-gitlab

### env configurations

- LGTM_DB_PATH string

  bolt db data (default "/var/lib/lgtm/lgtm.data")

- LGTM_GITLAB_URL string

  e.g. https://your.gitlab.com
      
- LGTM_COUNT int

  lgtm user count (default 2)

- LGTM_NOTE string

  lgtm note (default "lgtm")

- LGTM_LOG_LEVEL string

  log level (default "info")

- LGTM_PORT int

  http listen port (default 8989)

- LGTM_TOKEN string

  gitlab private token which used to accept merge request. can be found in https://your.gitlab.com/profile/account


### docker

```shell
docker run -d --restart=always \
    --name lgtm-gitlab \
    -e LGTM_TOKEN=YOUR_TOKEN \
    -e LGTM_GITLAB_URL=http://your_gitlab_url \
    -p 8989:8989 \
    cloverstd/lgtm-gitlab
```

## comment LGTM on MR

Now you can comment a LGTM on gitlab MR, when the `LGTM_COUNT` achieve, the MR will be merged.