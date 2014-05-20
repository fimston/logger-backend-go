
CREATE DATABASE logger
  WITH OWNER = postgres
       ENCODING = 'UTF8'
       TABLESPACE = pg_default
       LC_COLLATE = 'ru_RU.UTF-8'
       LC_CTYPE = 'ru_RU.UTF-8'
       CONNECTION LIMIT = -1;

create schema auth
authorization postgres;


create schema services
authorization postgres;

create domain services.cost as numeric
check (value >= 0.0);
--drop table services.accounts
create table services.accounts
(
	account_id bigserial primary key not null
	, title varchar not null
	, message_ttl interval not null
	, duration interval
	, price services.cost not null
	, traffic_per_day bigint
);



--drop table auth.users
create table auth.users
(
	user_id bigserial primary key not null
	, api_key varchar not null
	, name varchar not null
	, login varchar not null
	, passwd_md5 varchar not null
	, account_id bigint
	, account_expiration_date timestamp without time zone
	, constraint user_account_to_accounts_fk foreign key (account_id) references services.accounts	
);

CREATE UNIQUE INDEX user_api_key_uniq_idx  ON auth.users(api_key);
CREATE UNIQUE INDEX user_login_uniq_idx  ON auth.users(login);

create or replace function auth.generate_api_key(login varchar, passwd varchar)
returns varchar as
$$
	select md5(random()::varchar || now()::varchar || login || passwd);
$$
language sql cost 10;

create or replace function auth.create_user(name varchar, login varchar, passwd varchar)
returns bigint as
$$
	insert into auth.users
	(name, api_key, login, passwd_md5, account_id, account_expiration_date)
	select
		name, auth.generate_api_key(login, md5(passwd)), login, md5(passwd), account_id, now() + duration
	from services.accounts where account_id = 1
	returning user_id;
$$
language sql cost 10;

CREATE OR REPLACE FUNCTION services.interval_to_ms(interval)
  RETURNS bigint AS
  $$
	select (extract(epoch from $1) * 1000)::bigint
$$
  LANGUAGE sql VOLATILE
  COST 10;

create or replace function auth.select_users(
	out user_id bigint
	, out api_key varchar
	, out account_expiration_date timestamp without time zone
	, out message_ttl bigint)
returns setof record as
$$
	select user_id, api_key,  account_expiration_date, services.interval_to_ms(message_ttl)
	from auth.users
	left join services.accounts using (account_id)
$$
language sql cost 1000000;


/*



insert into services.accounts
values
(1, 'Тестовый', '1 day', null, 0.0, 100 * 1024 * 1024);


select auth.create_user('user1', 'login1', 'passwd1');
select auth.create_user('user2', 'login2', 'passwd2');
select auth.create_user('user3', 'login3', 'passwd3');
select auth.create_user('user4', 'login4', 'passwd4');

select * from auth.select_users()

*/