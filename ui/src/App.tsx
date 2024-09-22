import {createTheme, ThemeProvider} from "@mui/material";
import DataViewer from "./DataViewer";


const App = () => {

    const darkTheme = createTheme({
        palette: {
            mode: 'dark',
        },
    });

    return (
        <ThemeProvider theme={darkTheme}>
            <DataViewer/>
        </ThemeProvider>
    );
}

export default App
