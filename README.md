# anyquery-lastfm

An [Anyquery](https://anyquery.dev) plugin to query your Last.fm listening history and charts with SQL.

## Tables

| Table | Description | Required parameter |
|---|---|---|
| `recent_tracks` | Recently scrobbled tracks | `user` |
| `top_artists` | Top artists by play count | `user` |
| `top_tracks` | Top tracks by play count | `user` |
| `top_albums` | Top albums by play count | `user` |

### `recent_tracks`

```sql
SELECT * FROM anyquery_lastfm_recent_tracks('myusername') LIMIT 20;
```

Columns:

| Column | Type | Description |
|---|---|---|
| `artist` | TEXT | Track artist |
| `name` | TEXT | Track title |
| `album` | TEXT | Album title |
| `url` | TEXT | Last.fm track URL |
| `played_at` | TEXT | UTC scrobble time (RFC3339) |
| `now_playing` | BOOLEAN | Currently playing |
| `loved` | BOOLEAN | Loved by the user |
| `mbid` | TEXT | MusicBrainz ID |
| `image` | TEXT | URL of the largest available image |

### `top_artists`

```sql
SELECT * FROM anyquery_lastfm_top_artists('myusername') LIMIT 20;
```

Columns: `rank`, `name`, `playcount`, `listeners`, `url`, `mbid`, `image`.

### `top_tracks`

```sql
SELECT * FROM anyquery_lastfm_top_tracks('myusername') LIMIT 20;
```

Columns: `rank`, `name`, `artist`, `playcount`, `listeners`, `url`, `mbid`, `image`.

### `top_albums`

```sql
SELECT * FROM anyquery_lastfm_top_albums('myusername') LIMIT 20;
```

Columns: `rank`, `name`, `artist`, `playcount`, `url`, `mbid`, `image`.

## Configuration

You need a Last.fm API key. Get one at https://www.last.fm/api/account/create.

### Local development

Edit `devManifest.json` and add your API key:

```json
{
  "executable": "anyquery_lastfm.out",
  "build_command": "make",
  "user_config": {
    "default": {
      "api_key": "YOUR_LASTFM_API_KEY"
    }
  },
  ...
}
```

Then load the plugin in dev mode:

```bash
anyquery --dev -q "SELECT load_dev_plugin('anyquery_lastfm', 'devManifest.json');"
```

### Installed from registry

```bash
anyquery plugin install anyquery_lastfm
```

Anyquery will prompt you for the `api_key` during installation.

## Examples

```sql
-- Recent scrobbles
SELECT artist, name, played_at
FROM anyquery_lastfm_recent_tracks('myusername')
ORDER BY played_at DESC
LIMIT 10;

-- Scrobbles in the last 7 days
SELECT *
FROM anyquery_lastfm_recent_tracks('myusername')
WHERE datetime(played_at) > datetime('now', '-7 days');

-- Top artists
SELECT name, playcount
FROM anyquery_lastfm_top_artists('myusername')
LIMIT 10;

-- Export all scrobbles to a local table
CREATE TABLE my_scrobbles AS
SELECT * FROM anyquery_lastfm_recent_tracks('myusername');
```

## Pagination and rate limits

The plugin fetches 200 rows per page and automatically paginates until Last.fm has no more data. If you specify a `LIMIT`, it stops early.

Large exports (e.g. all scrobbles for a user with a long history) make many API calls and may be slow or hit Last.fm rate limits. Consider materializing the data with `CREATE TABLE AS SELECT` instead of querying live repeatedly.

## Development

Requirements:

- Go 1.22+
- Anyquery 0.4+

Build:

```bash
make
```

Run tests:

```bash
go test ./...
```

Load in Anyquery dev mode:

```bash
anyquery --dev
# then inside the shell:
SELECT load_dev_plugin('anyquery_lastfm', 'devManifest.json');
```

Reload after changes:

```sql
SELECT reload_dev_plugin('anyquery_lastfm');
```

## License

MIT
