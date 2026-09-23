# tknz 🎟️

Authentication for humans and bots.

## Running locally

### KMS
```sh
# create an ec private key to use with local kms
mkdir .tmp
openssl ecparam -name prime256v1 -genkey -noout -out .tmp/key.pem
```

```sh
# convert private key to a jwk (to use when adding jwks to the db)
go run ./cmd/mkjwks -keys .tmp/key.pem \
  | jq -c '.[0] + {kid: "local", use: "sig", key_ops: ["sign", "verify"]}'
```

```json
// e.g.
{
  "alg": "ES256",
  "kty": "EC",
  "crv": "P-256",
  "x": "…",
  "y": "…",
  "kid": "local",
  "use": "sig",
  "key_ops": ["sign", "verify"]
}
```

### DB

```sql
-- create 'sys' space
insert into spaces (id, slug, attrs) values ('sys', 'sys', '{}');

update spaces
set
  attrs = jsonb_set(attrs, '{_rev}',
  to_jsonb(extract(epoch from clock_timestamp())::integer))
where id = 'sys';


-- add jwk (replace <jwk> with json from private key)
insert into jwks (id, space_id, attrs, params) values (
  'local',
  'sys',
  '{}',
  '<jwk>'
);
```
