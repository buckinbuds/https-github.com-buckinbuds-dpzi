https://github.com/vvoluom/LinkPal
LinkPal-master>LinkPal-master>adapter
**MODIFY FILE PRIOR IMPLEMENTATION**

Origin file labeled type 'FILE'.

FROM node:alpine

WORKDIR /paypal-adapter
ADD . .

RUN apk add --no-cache git
RUN yarn install
RUN yarn build

ENV EA_PORT=8080

CMD node ./dist/server.js

