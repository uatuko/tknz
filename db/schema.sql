create table if not exists spaces (
	id    text not null,
	slug  text,
	attrs jsonb,

	constraint "spaces.pkey" primary key (id),
	constraint "spaces.unique" unique (slug),
	constraint "spaces.check-slug" check (slug <> ''),
	constraint "spaces.check-attrs" check (jsonb_typeof(attrs) = 'object')
);

create table if not exists apps (
	id        text not null,
	space_id  text not null,
	client_id text,
	attrs     jsonb,

	constraint "apps.pkey" primary key (id),
	constraint "apps.fkey-space_id" foreign key (space_id)
		references spaces(id)
		on delete cascade,
	constraint "apps.unique" unique (client_id),
	constraint "apps.check-attrs" check (jsonb_typeof(attrs) = 'object')
);
create index if not exists "apps.idx-space_id" on apps using hash (space_id);

create table if not exists federations (
	id     text not null,
	app_id text not null,
	iss    text not null,
	attrs  jsonb,

	constraint "federations.pkey" primary key (id),
	constraint "federations.fkey-app_id" foreign key (app_id)
		references apps(id)
		on delete cascade,
	constraint "federations.unique" unique (app_id, iss),
	constraint "federations.check-attrs" check (jsonb_typeof(attrs) = 'object')
);
create index if not exists "federations.idx-app_id" on federations using hash (app_id);

create table if not exists providers (
	id     text not null,
	app_id text not null,
	slug   text not null,
	attrs  jsonb,

	constraint "providers.pkey" primary key (id),
	constraint "providers.fkey-app_id" foreign key (app_id)
		references apps(id)
		on delete cascade,
	constraint "providers.unique" unique (app_id, slug),
	constraint "providers.check-attrs" check (jsonb_typeof(attrs) = 'object')
);
create index if not exists "providers.idx-app_id" on providers using hash (app_id);

create table if not exists idns (
	id      text not null,
	app_id  text not null,
	login   text not null,

	federation_id text,
	attrs         jsonb,

	constraint "idns.pkey" primary key (id),
	constraint "idns.fkey-app_id" foreign key (app_id)
		references apps(id)
		on delete cascade,
	constraint "idns.fkey-federation_id" foreign key (federation_id)
		references federations(id)
		on delete cascade,
	constraint "idns.unique" unique (app_id, login),
	constraint "idns.check-attrs" check (jsonb_typeof(attrs) = 'object')
);
create index if not exists "idns.idx-app_id" on idns using hash (app_id);
create index if not exists "idns.idx-federation_id" on idns using hash (federation_id);

create table if not exists idn_srcs (
	idn_id      text not null,
	provider_id text not null,
	sub         text not null,
	attrs       jsonb,

	constraint "idn_srcs.pkey" primary key (idn_id, provider_id),
	constraint "idn_srcs.fkey-idn_id" foreign key (idn_id)
		references idns(id)
		on delete cascade,
	constraint "idn_srcs.fkey-provider_id" foreign key (provider_id)
		references providers(id)
		on delete cascade,
	constraint "idn_srcs.unique" unique (provider_id, sub),
	constraint "idn_srcs.check-attrs" check (jsonb_typeof(attrs) = 'object')
);

create table if not exists jwks (
	id       text not null,
	space_id text not null,
	attrs    jsonb,
	params   jsonb,

	constraint "jwks.pkey" primary key (id),
	constraint "jwks.fkey-space_id" foreign key (space_id)
		references spaces(id)
		on delete cascade,
	constraint "jwks.check-attrs" check (jsonb_typeof(attrs) = 'object'),
	constraint "jwks.check-params" check (jsonb_typeof(params) = 'object')
);
create index if not exists "jwks.idx-space_id" on jwks using hash (space_id);

create table if not exists tokens (
	id    text      not null,
	exp   timestamp not null,
	attrs jsonb,

	constraint "tokens.pkey" primary key (id),
	constraint "tokens.check-attrs" check (jsonb_typeof(attrs) = 'object')
);
