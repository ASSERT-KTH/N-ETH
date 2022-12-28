import pandas as pd
import re
from scipy.stats import shapiro
import statistics

# pd.set_option('display.max_colwidth', None)
pd.options.display.max_rows = None 
pd.set_option("expand_frame_repr", False)

read_tcp_regex = r"(error: Post \"http://localhost:8545\":) read tcp 127.0.0.1:(\d+)->127.0.0.1:8545: (read: connection reset by peer)"

def compute_percentage(row):
    return row[3]/row[1]

def aggregate_read_tcp_error(row):
    r = re.sub(read_tcp_regex, r'\1 \3', row[2])
    return r

def normality_test(col):
    _, p = shapiro(col)
    return p

def std_dev_test(col):
    return statistics.stdev(col)
    

def availability_per_method():

    targets = ['geth', 'besu', 'erigon', 'nethermind']
    strats = [[22,24,28], [27,29,25], [24, 30, 26], [22, 23, 26]]
    full_df = pd.DataFrame()

    for i in range(len(targets)):
        target = targets[i]
        target_df = pd.DataFrame()

        for j in range(3):
            x = strats[i][j]
            df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses_random.dat', names=['idx', 'method_name', 'result'])
            df['result'] = df.apply(aggregate_read_tcp_error, axis=1)
            df1 = df.groupby(['method_name'])['idx'].size().reset_index(name='count')
            df1['total'] = df1['count']
            df1 = df1.drop('count', axis=1)
            
            df2 = df.groupby(['method_name', 'result']).count().reset_index(['result','method_name'])
            df2['count'] = df2['idx']
            df2 = df2.drop('idx', axis=1)


            df3 = df1.merge(df2, left_on='method_name', right_on='method_name')
            df3[f'{target}_{x}'] = df3.apply(compute_percentage, axis=1)

            df4 = df3[df3['result'] == ' success']
            df4 = df4.drop(['method_name', 'total', 'result', 'count'], axis=1).reset_index(drop=True)
            # df4.reset_index()

            target_df = pd.concat([target_df, df4], axis=1)
            target_df = target_df.fillna(0)
        full_df = pd.concat([full_df, target_df], axis=1)
        # full_df = pd.merge(full_df, target_df, left_index=True, right_index=True, how='outer')
    return full_df
    

def copmute_stdevs():

    targets = ['geth', 'besu', 'erigon', 'nethermind']
    strats = [[22,24,28], [27,29,25], [24, 30, 26], [22, 23, 26]]
    full_df = pd.DataFrame()

    for i in range(len(targets)):
        target = targets[i]
        target_df = pd.DataFrame()

        for j in range(3):
            x = strats[i][j]
            df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses_random.dat', names=['idx', 'method_name', 'result'])
            df['result'] = df.apply(aggregate_read_tcp_error, axis=1)
            df1 = df.groupby(['method_name'])['idx'].size().reset_index(name='count')
            df1['total'] = df1['count']
            df1 = df1.drop('count', axis=1)
            
            df2 = df.groupby(['method_name', 'result']).count().reset_index(['result','method_name'])
            df2['count'] = df2['idx']
            df2 = df2.drop('idx', axis=1)


            df3 = df1.merge(df2, left_on='method_name', right_on='method_name')
            df3[f'{target}_{x}'] = df3.apply(compute_percentage, axis=1)

            df4 = df3[df3['result'] == ' success']
            df4 = df4.drop(['method_name', 'total', 'result', 'count'], axis=1).reset_index(drop=True)

            target_df = pd.concat([target_df, df4], axis=1)
            target_df = target_df.fillna(0)

            # table = df4.to_markdown()
            # print(table)

            # f = open(f'./{target}/availability_per_method/error_model_{x}.md', "w")
            # f.write(table)
            # f.close()
        df5 = target_df.apply(std_dev_test).to_frame()
        df5 = df5.reset_index()
        df5.columns = ['target_fault_model', 'stdev']
        df5 = df5.sort_values(by=['stdev'], ascending=False)
        df5 = df5.head(3)

        full_df = pd.concat([full_df, df5])
        
    
    full_df = full_df.set_index('target_fault_model')
    full_df = full_df.transpose()
    return full_df

values = availability_per_method()
stdev = copmute_stdevs()

pd.set_option("display.precision", 3)
final = pd.concat([values,stdev])

print(final.to_latex())