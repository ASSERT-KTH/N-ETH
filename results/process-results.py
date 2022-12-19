import pandas as pd
import re

# pd.set_option('display.max_colwidth', None)  
read_tcp_regex = r"(error: Post \"http://localhost:8545\":) read tcp 127.0.0.1:(\d+)->127.0.0.1:8545: (read: connection reset by peer)"

def fn1(row):
    return row[3]/row[1]

def fn(row):
    r = re.sub(read_tcp_regex, r'\1 \3', row[2])
    return r

for x in range(1, 31):
    df = pd.read_csv(f'./geth/output/output-error_models_{x}_1.05/responses_random.dat', names=['idx', 'method_name', 'result'])
    df['result'] = df.apply(fn, axis=1)
    df1 = df.groupby(['method_name'])['idx'].size().reset_index(name='count')
    df1['total'] = df1['count']
    df1 = df1.drop('count', axis=1)
    
    df2 = df.groupby(['method_name', 'result']).count().reset_index(['result','method_name'])
    df2['count'] = df2['idx']
    df2 = df2.drop('idx', axis=1)

    df3 = df1.merge(df2, left_on='method_name', right_on='method_name')
    df3['percentage'] = df3.apply(fn1, axis=1)

    df4 = df3.loc[df3['result'] == ' success'].reset_index()
    df4 = df4.drop(['index'], axis=1)


    table = df4.to_markdown()

    f = open(f'./geth/availability_per_method/error_model_{x}.md', "w")
    f.write(table)
    f.close()


    