import { socketManager } from "../../utils/websocket/socket";

const LobbyPage = () => {

    async function connectToServer() {
        await socketManager.send("join_queue");
    }

    return (
        <div className="flex flex-col items-center justify-center h-screen">
            <h1 className="text-4xl font-bold mb-8">Lobby</h1>
            <p className="text-lg text-gray-600">Welcome to the lobby! Here you can join or create a game.</p>
            <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded" onClick={connectToServer}>
                Join Game
            </button>
        </div>
    );
}

export default LobbyPage;