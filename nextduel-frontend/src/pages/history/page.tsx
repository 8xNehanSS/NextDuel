import { useEffect, useState } from "react"
import { getDuelHistory } from "../../services/api/history"

const HistoryPage = () => {
    const [history, setHistory] = useState<any[]>([])

    useEffect(() => {
        loadData()
    }, [])

    const loadData = async () => {
        try {
            const response = await getDuelHistory()
            setHistory(response)
        } catch (error) {
            console.error("Error fetching duel history:", error)
        }
    }
return (
    <div className="w-6xl mx-auto p-6">
        <div className="space-y-4">
            {history.map((item, index) => (
                <div
                    key={index}
                    className="bg-white rounded-xl shadow-md border border-gray-200 p-5 hover:shadow-lg transition"
                >
                    <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-2">
                        <div>
                            <h2 className="text-lg font-semibold">
                                Match #{item.match_id}
                            </h2>
                            <p className="text-gray-600">
                                Player ID: {item.player_id}
                            </p>
                        </div>

                        <div className="flex gap-2">
                            <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                                Slot {item.slot}
                            </span>

                            <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm">
                                {item.rounds_won} Rounds Won
                            </span>

                            <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm">
                                Score {item.score}
                            </span>
                        </div>
                    </div>

                    <div className="flex flex-row gap-4 mt-4 text-sm text-gray-500">
                        <div>
                            <span className="font-medium text-gray-700">
                                Joined:
                            </span>{" "}
                            {new Date(item.joined_at).toLocaleString()}
                        </div>

                        <div>
                            <span className="font-medium text-gray-700">
                                Left:
                            </span>{" "}
                            {new Date(item.left_at).toLocaleString()}
                        </div>
                    </div>
                </div>
            ))}
        </div>

        {history.length === 0 && (
            <div className="text-center py-10 text-gray-500">
                No match history found.
            </div>
        )}
    </div>
);
}

export default HistoryPage