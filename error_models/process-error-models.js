const fs = require('fs')

const targets = ['besu', 'geth', 'erigon', 'nethermind']

const threshold = 0
const errorModel = './<target>/error_models.json'
const dest = './common/error_models_<topn>_<factor>.json'


const union = {}

const errormodels = targets.map(target => {
    const file = fs.readFileSync(errorModel.replace('<target>', target))
    return JSON.parse(file).experiments
})

errormodels.forEach(em => {
    em.forEach(tuple => {
        key = `${tuple.syscall_name}${tuple.error_code}`
        if (!union[key]) {
            union[key] = []
        }
        union[key].push(tuple)
    })
})

const filtered = []

Object.keys(union).forEach(k => {
    if (union[k][0].syscall_name !== "accept4") {
        filtered.push(union[k])
    }
})

const minRates = filtered.map(em => {

    var min = 99
    var max = -1

    em.forEach(x => {
        min = Math.min(x.original_mean_rate, min)
        max = Math.max(x.original_mean_rate, max)
    })

    return {
        syscall_name: em[0].syscall_name,
        error_code: em[0].error_code,
        original_mean_rate: min
        // original_mean_rate_max: max
    }
})

console.log(minRates.length)
const aggroFactor = [1.05]
const aggroTopN = Array.from({length: 30}, (_, i) => i + 1)

const main = () => {

    aggroTopN.forEach(n => {

        aggroFactor.forEach(factor => {
            const processed = minRates
                .filter(model => model.original_mean_rate > threshold && model.original_mean_rate < 1)
                .sort((a, b) => b.original_mean_rate - a.original_mean_rate)
                .map(obj => {
                    return {
                        syscall_name: obj.syscall_name,
                        error_code: obj.error_code,
                        original_mean_rate: obj.original_mean_rate,
                        failure_rate: obj.original_mean_rate * factor
                    }
                })
                .slice(0, n)

            const obj = {
                experiments: processed
            }
            fs.writeFileSync(dest.replace('<factor>', `${factor}`).replace('<topn>', `${n}`), JSON.stringify(obj, null, 2))
        })
    })
}

main()
