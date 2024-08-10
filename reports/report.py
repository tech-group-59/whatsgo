from typing import Optional, Tuple

import pandas as pd
import sqlite3

# database has table `messages` with columns:
# - id
# - sender
# - chat
# - content
# - timestamp
# - parsed_content
sqlite_path = 'prod_whatsgo_07_08_2024.db'

# need to work only with one chat
chat = '120363311602503571@g.us'

# keywords to search in messages
# keyword file has two columns: `keyword` and `location`
keywords_path = 'keywords.csv'


def read_needles():
    """
    Read needles from file and return dictionary with two sets: `keyword` and `location`
    All words are lowercased
    Filter out empty strings and nan values
    """
    df = pd.read_csv(keywords_path)
    df = df.apply(lambda x: x.str.lower())
    df = df.apply(lambda x: x.str.strip())
    needles = {
        'keyword': set(df['keyword'].dropna()),
        'location': set(df['location'].dropna())
    }
    return needles


blacklist = [
    'горловка',
]


def check_text(text, needles) -> Optional[Tuple[str, str]]:
    """
    Check if text contains any of the keyword and any of the location
    and check if text does not contain any of the blacklist words
    """
    text = text.lower()
    for keyword in needles['keyword']:
        if keyword in text:
            for location in needles['location']:
                if location in text:
                    for word in blacklist:
                        if word in text:
                            return None
                    return keyword, location
    return None


def get_data():
    # connect to database
    con = sqlite3.connect(sqlite_path)
    query = f'SELECT *  FROM messages  WHERE timestamp BETWEEN "2024-07-28" AND "2024-08-07" AND chat = "{chat}"'
    df = pd.read_sql(query, con)
    con.close()
    return df


if __name__ == '__main__':
    needles = read_needles()

    data = get_data()

    result_data = []

    for _, row in data.iterrows():
        result = check_text(row['content'], needles)
        if result:
            result_row = {
                'location': result[1],
                'keyword': result[0],
                'timestamp': row['timestamp'],
                'content': row['content']
            }
            print(row['timestamp'], row['sender'], result)
            result_data.append(result_row)

    # result df should have columns: location, keyword, timestamp, content
    result_df = pd.DataFrame(result_data)


    # timestamp is in a format like `2024-07-28 00:02:05 +0300 EEST`
    # Function to remove the timezone name
    def remove_timezone_name(date_str):
        return ' '.join(date_str.split()[:-1])


    # Apply the function to the 'date' column
    result_df['timestamp'] = result_df['timestamp'].apply(remove_timezone_name)
    # Parse the modified date strings
    result_df['timestamp'] = pd.to_datetime(result_df['timestamp'], format='%Y-%m-%d %H:%M:%S %z')

    result_df['date'] = result_df['timestamp'].dt.date
    result_df['time'] = result_df['timestamp'].dt.time

    # remove timestamp column
    # result_df = result_df.drop(columns=['timestamp'])

    result_df.to_csv('result.csv', index=False)

    # 1. aggregate data by location and date
    result_df1 = result_df.groupby(['location', 'date']).size().reset_index(name='count')
    result_df1.to_csv('result_aggregated_location_date.csv', index=False)

    # 2. aggregate data by location
    result_df2 = result_df.groupby(['location']).size().reset_index(name='count')
    result_df2.to_csv('result_aggregated_location.csv', index=False)
