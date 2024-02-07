# doc_drivers
WIP. Utility that generates drivers for [hackborn/doc](https://github.com/hackborn/doc)

## Use
This utility converts a collection of domain classes into a driver for [hackborn/doc](https://github.com/hackborn/doc), allowing use of the doc package to store and load data to a database backend. For instructions on how to use the resulting driver see that project; this one is concerned with generating the driver.

## Tags
Translation to a database can be customized through the use of `doc` field tags. By default, every field in a struct has a corresponding field in the database with the same name, but this can be modified.

### Tag Name Rules
1. If there is no tag, the struct field name is used as the database field name.
```
Name string
```
The database field name will be `Name`.

2. The first string in the `doc` tag is used as the database name.
```
Name string `doc:"name"`
```
The database field name will be `name`.

3. Tags are comma separated; if you omit a first string, the struct field name is used for the database field name.
```
Name string `doc:","`
```
The database field name will be `Name`.

4. A tag name of `-` will omit the field from the database.
```
Name string `doc:"-"`
```
The struct Name field will have no corresponding database field.

5. Unexported fields with no tag are ignored.
```
name string
```
The struct name field will have no corresponding database field.

6. Unexported fields with a tag are treated as table tags.
```
_table string `doc:"company"`
```
The database table will be named `company`.

### Tag Key Rules
1. Database keys are specified by using the `key` keyword.
```
Id int `doc:"id, key"`
```
The database will have a key of `id`.

2. Multiple keys can be specified for compound database keys.
```
Pri int `doc:"pri, key"`
Sec int `doc:"sec, key"`
```
The database will have a compound key of `pri, sec`.

3. Groups of keys can be created by providing key names.
```
Id int `doc:"id, key"`
Name string `doc:"name, key(compound)"`
Date int `doc:"date, key(compound)"`
```
The database will have keys of `id` and `name, date`.

4. Keys within a key group can be ordered by supplying a number after the name.
```
Name string `doc:"name, key(compound, 1)"`
Date int `doc:"date, key(compound, 0)"`
```
The database will have a key of `date, name`.

5. Key rules vary depending on the underlying storage model and level of support in the driver. Currently the only special rule is that the "primary" key is the first key group name, alphabetically. The easiest way to specify a primary key is to leave the key name blank.