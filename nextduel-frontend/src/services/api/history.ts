import { api } from "../../utils/apiclient";

export async function getDuelHistory(): Promise<any[]> {
    const response = await api.get("/history/duels");
    return response as any[];
}