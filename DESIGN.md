# Database for turbo

A sane default when it come to DB is postgres, and even if it not the best choice, it will most likely
perform well for few million messages. But before making that decision we want to consider
if there are better options available for the chat application we are planing to build.

Most live message send will not involve a read, just a write.
A live chat session might require one read to get previous messages, and then write for every message. 
We can consider using redis for caching messages and bulk writing it in the DB. But I feel it is
safe to say this is more write heavy than application like an email service. Most blogs an online
resources seem to agree to this, stating read/write ration is 1:1.
We should think of ways of measuring thas.

---

## Options to consider
 - Posgres
	 - Pro:
		- We have used it all our dev lives, so less learning and fast development.
	    - I like elephants
	- Cons:
		- Usually not recommended for write heavy load.
	    -  Might have to add a cache layer to have decent perf, if we get reasonable # of users. 
 - MongoDB
	 - Pro:
		 - No schema so we can dump whatever we want.( Read as fast iteration)
	- Cons:
		- Me no like it
	    - Boooo
 - Cassandra
	 - Pros:
		 - Better for fast writes.
		 - Table like schema, is available.
	- Cons:
		- Don't know any thing about using this.
		- It is an AP DB. Dealing with this will require some time and effort.
 - Scylla DB
	- Pros:
		- Regularly required cleanups are faster. 
		- Advertised to be better than Cassandra
	- Cons:
		- All of cassandra
		- Know even lesser about this
 -  HBase
	 - Pros:
		 - Favors consistency Consistent
		 - Marketed as Open source Google Big Table

---

### Key takeaway: 

With HBase and PG, the whole database can go down should the master node fail. With Cassandra, Scylla, on the other hand, if a node goes down the database will still be available. 
However, because of the masterless architecture, data inconsistencies can occur.

So essentially the Question is CP v AP.

If we want to be flexible and be able to switch DBs we definitely need to make sure DB call are abstracted away from business logic. And we can then pug different DB adapters as we please.

---

