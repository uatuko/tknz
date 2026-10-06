# Svelte

## Setup

```sh
# generate ec p-256 private key
if [ ! -d .tmp ]; then mkdir .tmp; fi
openssl ecparam -name prime256v1 -genkey -noout -out .tmp/ecdsa_p-256.pem

# copy pkcs#8 formatted key (to update .env.local)
openssl pkey -in .tmp/ecdsa_p-256.pem | pbcopy

# generate jwks (to insert into db)
go run ../../cmd/mkjwks -keys .tmp/ecdsa_p-256.pem
```

```sql
-- space (require sys space)
insert into spaces (id, slug, attrs) values ('local', 'local', '{}');
update spaces
set
  attrs = jsonb_set(attrs, '{_rev}',
  to_jsonb(extract(epoch from clock_timestamp())::integer))
where id = 'local';

-- app
insert into apps (id, space_id, client_id, attrs)
values (
  'local',
  'local',
  'local',
  '{"aud": "http://localhost:5173", "redirect_uris": ["http://localhost:5173/sign-in"]}'
);

update apps
set
  attrs = jsonb_set(attrs, '{_rev}',
  to_jsonb(extract(epoch from clock_timestamp())::integer))
where id = 'local';

update apps
set attrs = attrs || '{"keys":[]}'::jsonb -- jwks generated from ec p-256 private key
where id = 'local';

-- providers
insert into providers (id, app_id, slug, attrs)
values (
  'local-password',
  'local',
  'password',
  '{}'
);

update providers
set
  attrs = jsonb_set(attrs, '{_rev}',
  to_jsonb(extract(epoch from clock_timestamp())::integer))
where id = 'local-password';

-- user (with password sign-in)
insert into idns (id, app_id, login, attrs)
values (
  '06gfo7f5bhuatca9jt4ha1hp7g', -- UUIDv7
  'local',
  'user',
  '{"email": "user@example.local"}'
);

update idns set attrs['_rev'] = to_jsonb(extract(epoch from clock_timestamp())::integer)
where id = '06gfo7f5bhuatca9jt4ha1hp7g';

insert into idn_srcs (idn_id, provider_id, sub, attrs)
values (
  '06gfo7f5bhuatca9jt4ha1hp7g',
  'local-password',
  'user',
  '{"pwd": {
    "typ":"argon2",
    "salt":"xzx5ISdtFChBP871mnc8nw==",
    "key":"0Qm8aK/kBOqoX5RcwkJBGGRlLOlVTcnzdJIJDPSwfhY=",
    "time":3,
    "memory":65536,
    "threads":4
  }}' -- password = 'pass'
);

update idn_srcs set attrs['_rev'] = to_jsonb(extract(epoch from clock_timestamp())::integer)
where idn_id = '06gfo7f5bhuatca9jt4ha1hp7g' and provider_id = 'local-password';
```

## Convert *.proto files to JSON descriptors

```sh
node cmd/mkpb.js \
  --protodir=../../proto \
  --out=./src/lib/server/proto/authn.json \
  tknz/v1/authn.proto
```

## Generate protobuf types

```sh
npx proto-loader-gen-types \
  --defaults \
  --oneofs \
  --grpcLib=@grpc/grpc-js \
  --includeDirs=../../proto \
  --outDir=./src/lib/server/proto/ \
  tknz/v1/authn.proto
```

To fix `Original file:` paths,

```
sed -i '' 's/\.\.\/\.\.\/proto\///g' src/lib/server/proto/tknz/v1/*.ts
```
