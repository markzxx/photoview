cd /photoview/api
git pull
go build -v -o photoview .
cd /photoview/ui
npm i
npm run build