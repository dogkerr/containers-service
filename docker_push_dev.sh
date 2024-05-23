 docker build .  -f Dockerfile.debug -t lintangbirdas/container-service-dev:v1
docker tag lintangbirdas/container-service-dev:v1 lintangbirdas/container-service-dev:v1
docker push lintangbirdas/container-service-dev:v1