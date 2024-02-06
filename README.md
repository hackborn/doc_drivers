# doc_drivers
WIP. Utility that generates drivers for [hackborn/doc](https://github.com/hackborn/doc)

## Use
This utility converts a collection of domain classes into a driver for [hackborn/doc](https://github.com/hackborn/doc), allowing use of the doc package to store and load data to a database backend. For instructions on how to use the resulting driver see that project; this one is concerned with generating the driver.

## Tags
Translation to a databse can be customized through the use of `doc` field tags. By default, every field in a struct has a corresponding field in the database with the same name, but this can be modified.

### Tag Name
When no tag is present the struct field name will be used as the database field name. For example,
```
Name string
```
will use a database field named `Name`.

You can set an explicit database field name by supplying the name as the first element of a `doc` tag. For example,
```
Name string `doc:"name"`
```
will use a database field named `name`.

You can explicitly use the struct field name as the database name by omitting a name and including a separator. For example,
```
Name string `doc:","`
```
will use a database field named `Name`.

You can prevent the field from having a corresponding database field by using `-`:
```
Name string `doc:"-"`
```
will have no database field named.

### Tag Key
Keys are specified by using the `key` keyword after specifying the name (i.e., after a `,` in the struct tag). Keys optionally have a name and a position index. For example,
```
Id int `doc:"id, key"`
```
will have an unnamed key on the `id` field.

To supply a name for a key, use
```
Id int `doc:"id, key(c)"`
```

To give a key a position in its key group, use
```
Id int `doc:"id, key(c,0)"`
```

For example, to create a key named "primary" with two ordered keys, use
```
Partition string `doc:"partition, key(primary,0)"`
Sort string `doc:"sort, key(primary,1)"`
```

One thing to note, the "primary" key is the one with the first name, alphabetically. The easiest thing to do is leave the primary key unnamed.