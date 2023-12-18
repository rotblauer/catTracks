
This CLI command creates a copy of a `tracks.db` database, but with the new db containing only the `snaps` bucket of the tracks.
CatTracks APIs and the website depend on the availability of snaps for `api.catonmap.info/snaps`.

But the tracks.db is getting huge (355GB and counting).
So I want to backup this old, swollen database, and re-init with a new one, but
maintaining the functionality of the website.

Eventually I want to STOP SAVING the tracks in the "tracks" bucket of the database entirely, using only .json.gz files for persistent and general storage of tracks.
