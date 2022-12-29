import pandas as pd
import re

pd.set_option('display.max_colwidth', None)
pd.options.display.max_rows = None 
pd.set_option("expand_frame_repr", False)

read_tcp_regex = r"(error: Post \"http://localhost:8545\":) read tcp 127.0.0.1:(\d+)->127.0.0.1:8545: (read: connection reset by peer)"
http_malformed_regex = r"(error: Post \"http://localhost:8545\": net/http: HTTP/1.x transport connection broken: malformed HTTP response).*"


post_eof = r'(error: Post) "http://localhost:8545": (EOF)'
post_timeout = r'(error: Post) "http://localhost:8545": context deadline exceeded \((.*)\)'
post_timeout1 = r'(error:) context deadline exceeded \((.*)\)'
post_refused = r'(error: Post) "http://localhost:8545": dial tcp 127.0.0.1:8545: (connect: connection refused)'
post_http = r'(error: Post) "http://localhost:8545": net/http: HTTP/1.x transport connection broken: (malformed HTTP response)'
post_reset = r'(error: Post) "http://localhost:8545": (read: connection reset by peer)' 

def compute_percentage(row):
    return row[3]/row[1]

def aggregate_read_tcp_error(row):
    r = re.sub(read_tcp_regex, r'\1 \3', row[2])
    return r
def aggregate_bad_http_error(row):
    r = re.sub(http_malformed_regex, r'\1', row[2])
    return r

def postprocess(row):
    r = re.sub(post_eof, r'\1: \2', row[0])
    r = re.sub(post_timeout, r'\1: \2', r)
    r = re.sub(post_timeout1, r'\1: \2', r)
    r = re.sub(post_refused, r'\1: \2', r)
    r = re.sub(post_http, r'\1: \2', r)
    r = re.sub(post_reset, r'\1: \2', r)
    return r

def availability_per_method():

    targets = ['geth', 'besu', 'erigon', 'nethermind']
    # targets = ['besu']
    full_df = pd.DataFrame()
    full_df['error'] = 0

    for i in range(len(targets)):
        target = targets[i]
        target_df = pd.DataFrame()

        for x in range(1, 31):
            df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses_random.dat', names=['idx', 'method_name', 'error'])
            df['error'] = df.apply(aggregate_read_tcp_error, axis=1)
            df['error'] = df.apply(aggregate_bad_http_error, axis=1)

            df1 = df[df['error'] != ' success']
            df1 = df1.drop(['idx', 'method_name'], axis=1)

            df2 = df1.drop_duplicates()
            df2 = df2.reset_index(drop=True)
            
            target_df = pd.concat([target_df, df2])
            
        target_df = target_df.drop_duplicates()
        target_df = target_df.set_index('error', drop=True)
        target_df['value'] = 1
        full_df = full_df.join(target_df, how='outer', rsuffix=target)

    full_df = full_df.drop('error', axis=1)
    full_df = full_df.fillna(0)
    
    full_df = full_df.reset_index()
    full_df['error'] = full_df.apply(postprocess, axis=1)

    full_df = full_df.set_index('error', drop=True)

    return full_df
    

pd.set_option("display.precision", 3)
r = availability_per_method()

print(r.to_latex())
