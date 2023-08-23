const fs = require('fs')


const main = ()  => {

    final = []

    //for range 1 to 30
    for (let t = 1; t <= 30; t++) {
        for (let i = 1; i <= t; i++) {
            //load json file
            const json = require(`./error_models_${i}_1.05.json`)
            
            const error_list = json["experiments"]

            // get last item in error list
            const last_item = error_list[error_list.length - 1]

            //if last item's syscall_name is not repeated in final
            if (!final.some(item => item.syscall_name === last_item.syscall_name)) {
                //add to final  
                final.push(last_item)
            }
        }
        
        obj = {
            experiments: final
        }

        //print final to json file
        fs.writeFileSync(`./non-repeat/error_models_${final.length}_1.05.json`, JSON.stringify(obj, null, 2))
    }
}



main()