for /f "tokens=3" %%a in ('docker images -a ^| find "<none>"') do docker rmi %%a
docker-compose up --build
