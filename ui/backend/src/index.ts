import express, { Express, Request, Response } from 'express'
import cors from 'cors'
import {Pod} from 'kubernetes-types/core/v1'

const k8s = require('@kubernetes/client-node')

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sApi = kc.makeApiClient(k8s.CoreV1Api);
const k8sCRDApi = kc.makeApiClient(k8s.CustomObjectsApi)

const getPo = async (ns: string) => {
       const response =  await k8sApi.listNamespacedPod(ns)
       const podNames = response.body.items.map((pod : Pod) => ({
           name: pod?.metadata?.name,
           image: pod?.spec?.containers[0].image
        }))
       return podNames
};

const getPL = async (ns: string) => {
    const response =  await k8sCRDApi.listNamespacedCustomObject(
        'pipelines.bramble.dev',
        'v1alpha1',
        ns,
        'pipelines'
    )
    const pls = response.body.items.map((pl : any) => ({
        name: pl?.metadata?.name,
        tasks: pl?.spec?.tasks
    }))
    return pls
}

console.log(getPL('default'))

const app: Express = express()

const port: number = 5555

app.use(cors())
app.use(express.json())

app.get('/', async (req: Request, res: Response) => {
    try {
        const podNames = await getPL('default')
        res.json(podNames)
    } catch (error) {
        res.status(500).json({error: "internal server error (womp womp)"})
    }
})

app.listen(port, () => {
    console.log("Server is running")
})
