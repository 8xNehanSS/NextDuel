const Navigator = () => {
    return (
        <nav className="shadow-md px-6 py-4">
            <ul className="flex gap-6 justify-center text-gray-400 font-medium">
                <li>
                    <a
                        href="/"
                        className="hover:text-blue-500 transition-colors duration-200"
                    >
                        Home
                    </a>
                </li>
                <li>
                    <a
                        href="/lobby"
                        className="hover:text-blue-500 transition-colors duration-200"
                    >
                        Lobby
                    </a>
                </li>
                {/* <li>
                    <a
                        href="/contact"
                        className="hover:text-blue-500 transition-colors duration-200"
                    >
                        Contact
                    </a>
                </li> */}
            </ul>
        </nav>
    );
};

export default Navigator;