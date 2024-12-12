# sigmatest-kreditplus

Masuk ke directory project ini.

Masuk kedalam service user setup .env (copas dari example env)
Masuk kedalam service customer setup .env (copas dari example env)

Jalankan 
docker compose up -d 

Setting Minio
masuk kedalam [localhost:9000](http://localhost:9001/)
username dan password : minioadmin
lalu buat bucket bernama "sigmatech"
setting access key : http://localhost:9001/access-keys
lalu masukkan credentials (access_key dan secret_key kedalam env service customer) Create access.

Service customer akan running di localhost:9092
Service user akan running di localhost:9091
