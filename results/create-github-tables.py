import pandas as pd
import re
import ast

# pd.set_option('display.max_colwidth', None)  
read_tcp_regex = r"(error: Post \"http://localhost:8545\":) read tcp 127.0.0.1:(\d+)->127.0.0.1:8545: (read: connection reset by peer)"

def compute_percentage(row):
    return row[3]/row[1]

def aggregate_read_tcp_error(row):
    r = re.sub(read_tcp_regex, r'\1 \3', row[2])
    return r

def availability_per_method(target):
    for x in range(1, 31):
        df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses_random.dat', names=['idx', 'method_name', 'result'])
        df['result'] = df.apply(aggregate_read_tcp_error, axis=1)
        df1 = df.groupby(['method_name'])['idx'].size().reset_index(name='count')
        df1['total'] = df1['count']
        df1 = df1.drop('count', axis=1)
        
        df2 = df.groupby(['method_name', 'result']).count().reset_index(['result','method_name'])
        df2['count'] = df2['idx']
        df2 = df2.drop('idx', axis=1)

        df3 = df1.merge(df2, left_on='method_name', right_on='method_name')
        df3['percentage'] = df3.apply(compute_percentage, axis=1)

        table = df3.to_markdown()

        f = open(f'./{target}/availability_per_method/error_model_{x}.md', "w")
        f.write(table)
        f.close()


def hex_to_dec(row):
    return ast.literal_eval(row[2])

def clean_errors(row):
    if row[2].startswith(' error'):
        return '-0x1'
    return row[2]

def compute_distance(row):
    distance = row[3] - int(row[4])
    return distance

def compute_result(row):
    if row[4] == -1:
        return 'unavailable'
    elif row[5] > 5:
        return 'degraded'
    else:
        return 'available'


def compute_percentage_get_block(row):
    return int(row[0]) / 360_000

def availability_get_block(target):
    for x in range(1, 31):
        df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses-get-block.dat', names=['idx', 'method_name', 'target_block_hex', 'oracle_block'])
        df['target_block_hex'] = df.apply(clean_errors, axis=1)
        df['target_block'] = df.apply(hex_to_dec, axis=1)
        df['distance'] = df.apply(compute_distance, axis=1)
        df['status'] = df.apply(compute_result, axis=1)

        df1 = df.drop(['method_name', 'target_block_hex', 'oracle_block', 'target_block', 'distance'], axis=1)
        df2 = df1.groupby(['status']).count()

        df2['percentage'] = df2.apply(compute_percentage_get_block, axis=1)
        df2['count'] = df2['idx']
        df2 = df2.drop(['idx'], axis=1)

        print(df2)

        # f = open(f'./{target}/availability_per_method/error_model_{x}.md', "w")
        # f.write(table)
        # f.close()



availability_get_block('geth')