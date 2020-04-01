# gitdm

gitdm for the LF - human readable YAML file containing affiliations data


# Adding/Updating affiliation

If you find any errors or missing affiliations in those lists, please submit a pull request with edits to profile files: [1](https://github.com/LF-Engineering/gitdm/blob/master/profiles1.yaml), [2](https://github.com/LF-Engineering/gitdm/blob/master/profiles2.yaml), ... .


# YAML format


We are trying to keep that YAML as small as possible, so property names are a bit cryptic and only records containing at least one enrollment and at least one identity are exported:

This is the current format:


```
---
P:                                          # 'profiles' (profile holds possible multiple profile (called identities) from different data sources
                                            # for example git, GitHub, Jira, Slack etc.
- C: PL                                     # profile's 'country code' - if defined must be a correct two letter country code
  E: lukaszgryglicki!o2.pl                  # profile's 'email', all emails have their '@' replaced with '!'
  R:                                        # profile 'enrollments' list, at lease one enrollment must be present
  - T: "2006-03-01"                         # enrollment 'date to' - required
    C: Independent                          # enrollment 'organization' - required
    F: "1970-01-01"                         # enrollment 'date from' - required
(...)
  S: male                                   # profile's 'sex'/'gender'
  I:                                        # profiles 'identities' list (each profile must have source liek git/Jira//Slack), at least one identity must be present
                                            # each source can have multiple identities (multiple profiles, like for example multiple GitHub accounts)
  - E: lukaszgryglicki!o2.pl                # identity's 'email'
    M: Lukasz Gryglicki                     # identity's 'name'
    S: git                                  # identity's 'source' - required (the only required field)
(...)
  - E: lukaszgryglicki!o2.pl                # identity's 'email'
    M: Łukasz Gryglicki                     # identity's 'name'
    S: github                               # identity's 'source'
    U: lukaszgryglicki                      # identity's 'username' - some data sources have it (like GitHub), so doesn't (like git)
  B: 0                                      # profile's bot flag: 0 - normal profile, 1: bot profile
  U: Łukasz Gryglicki                       # profile's name
```

# Running locally

- To compile locally use: `make`.
- To run local sync service: `./serve.sh`.
- To do request to local service (check service used for reacting to PRs): `./check.sh`.
- To do request to local service (sync service used for reacting to push to master branch): `./sync.sh`.


# Docker

- Build docker image: `DOCKER_USER=... docker/build_image.sh`.
- Run it: `DOCKER_USER=... docker/run.sh`. It will serve on 17070 instead of 7070 port.
- Shell into the container: `DOCKER_USER=... docker/shell.sh`.
- Test request, `SYNC_URL` must be provided to specify non-default 17070 port: `SYNC_URL='127.0.0.1:17070' ./request.sh`.

# Kubernetes/Helm

To deploy on Kubernetes

- Go to `helm/`, run (LF real world example): `./setup.sh prod`.
- Eventually adjust Helm chart to your needs, including `setup.sh` and `delete.sh` shell scripts.
- Run from repository root directory (test env): `` SYNC_URL="`cat helm/gitdm/secrets/SYNC_URL.test.secret`" ./request.sh ``.
- Run from repository root directory (prod env): `` SYNC_URL="`cat helm/gitdm/secrets/SYNC_URL.prod.secret`" ./request.sh ``.

# GitHub actions

- Add you sync URL (for example AWS ELB of gitdm service stored in `helm/gitdm/secrets/SYNC_URL.test.secret`) in GitHub repository (Settings -> Secrets -> New secret: `SYNC_URL`).
- Configre actions in `.github/workflows/`, for example: `.github/workflows/push.yaml`.
