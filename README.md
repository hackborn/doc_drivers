# doc_drivers

WIP. Utility that generates drivers for [hackborn/doc](https://github.com/hackborn/doc)

## Use

This utility converts a collection of domain classes into a driver for [hackborn/doc](https://github.com/hackborn/doc), allowing use of the doc package to store and load data to a database backend. For instructions on how to use the resulting driver see that project; this one is concerned with generating the driver.

## Types

The goal of any driver is for all types to be transparently written to and read from the database. In practice, the default string, int and float types are handled natively, but more advanced types, like slices and maps, may need to be translated to a format the underlying database can handle. Ideally as a user you are left unware of this detail, but see the format tag below for details about manually forcing a translation.

## Tags

Translation to a database can be customized through the use of `doc` field tags. By default, every field in a struct has a corresponding field in the database with the same name, but this can be modified.

There are multiple keywords supported, all of which allow setting parameters in a `()` block.

Multiple keywords are separated with a `,`.

Unexported fields are a special case, setting rules for the table.

### Empty Tag

If there is no tag, the struct field name is used as the database field name. Note that some drivers might modify it for convention (for example, the SQLITE driver will lowercase the name).

```
Id string
```

The database field name will be `Id` or `id`, depending on the driver.

### Tag Keyword: Name

The `name` keyword allows directly setting the database name for a field.

1. Use a "name" property to directly set a databse name.

```
Id string `doc:"name(id)"`
```

The database field name will be `id`.

2. Unexported fields with no tag are ignored.

```
id string
```

The struct name field will have no corresponding database field.

3. Unexported fields with a tag are treated as table tags.

```
_table string `doc:"name(company)"`
```

The database table for the struct will be named `company`.

### Tag Keyword: Key

1. Database keys are specified by using the `key` keyword.

```
Id int `doc:"key"`
```

The database will have a key of `Id` or `id`.

2. Multiple keys can be specified for compound database keys.

```
Pri int `doc:"key"`
Sec int `doc:"key"`
```

The database will have a compound key of `Pri, Sec` or `pri, sec`.

3. Groups of keys can be created by providing key names.

```
Id int `doc:"key"`
Name string `doc:"key(group1)"`
Date int `doc:"key(group1)"`
```

The database will have keys of `id` and `name, date`.

4. Keys within a key group can be ordered by supplying a number after the name.

```
Name string `doc:"key(group1, 1)"`
Date int `doc:"key(group1, 0)"`
```

The database will have a key of `date, name`.

5. Key rules vary depending on the underlying storage model and level of support in the driver. Currently the only special rule is that the "primary" key is the first key group name, alphabetically. The easiest way to specify a primary key is to leave the key name blank.

### Tag Keyword: -

A tag of `-` will omit the field from the database.

```
Name string `doc:"-"`
```

The struct Name field will have no corresponding database field.

## Developing Drivers

The cmd/driverutil application is a tool used to help develop new drivers. Running the app displays a list of commands involved in generating the driver. See readmes for a specific driver (in backends/) for details.
