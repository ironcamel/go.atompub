
create table atom_feed (
    id text primary key,
    title text not null unique,
    title_type text not null default 'text',
    updated timestamp not null default now(),
    author_name text,
    author_email text,
    author_uri text
);

create table atom_entry (
    id text primary key,
    order_id serial unique,
    feed_title text not null,
    title text not null,
    title_type text not null default 'text',
    content text not null,
    content_type text not null default 'text',
    updated timestamp not null default now(),
    author_name text,
    author_email text,
    author_uri text
);

alter table atom_entry
add constraint atom_entry_feed_title_fkey
FOREIGN KEY (feed_title) REFERENCES atom_feed(title) ON UPDATE CASCADE;

create index atom_entry_feed_title_idx on atom_entry (feed_title);

create table atom_feed_link (
    id serial primary key,
    feed_id text not null,
    href text not null,
    rel text,
    hreflang text,
    title text,
    length int
);

alter table atom_feed_link
add constraint atom_feed_link_feed_id_fkey
FOREIGN KEY (feed_id) REFERENCES atom_feed(id)
on update CASCADE on delete CASCADE;

