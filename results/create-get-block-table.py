import pandas as pd
import ast

def hex_to_dec(row):
    # print(row[2])
    return ast.literal_eval(row[2])

def clean_errors(row):
    if row[2].startswith(' error') or row[2] == ' ':
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
    return int(row[1]) / 360_000

def availability_get_block():
    full_df = pd.DataFrame()
    targets = ['geth', 'besu', 'erigon', 'nethermind']
    for i in range(len(targets)):
        target = targets[i]
        target_df = pd.DataFrame()
        for x in range(1, 31):
            df = pd.read_csv(f'./{target}/output/output-error_models_{x}_1.05/responses-get-block.dat', names=['idx', 'method_name', 'target_block_hex', 'oracle_block'])
            df['target_block_hex'] = df.apply(clean_errors, axis=1)
            df['target_block'] = df.apply(hex_to_dec, axis=1)
            df['distance'] = df.apply(compute_distance, axis=1)
            df['status'] = df.apply(compute_result, axis=1)

            df1 = df.drop(['method_name', 'target_block_hex', 'oracle_block', 'target_block', 'distance'], axis=1)
            df2 = df1.groupby(['status']).count().reset_index()

            df2[target] = df2.apply(compute_percentage_get_block, axis=1)
            df2 = df2[df2['status'] == 'available']

                
            df2['strategy'] = f'FI{x}'
            df2 = df2.drop(['idx', 'status'], axis=1)
            if len(df2) == 0 :
                row = {
                    target: 0,
                    'strategy': f'FI{x}'
                }
                df2 = df2.append(row, ignore_index=True)
            
            if i != 0:
                df2 = df2.drop(['strategy'], axis=1)
            
            target_df = pd.concat([target_df, df2])
        full_df = pd.concat([full_df, target_df], axis=1)
    
    full_df = full_df.set_index('strategy', drop=True)
    return full_df

x = availability_get_block()

print(x.to_latex())
