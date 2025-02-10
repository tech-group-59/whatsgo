import { createUseStyles } from 'react-jss';
import { FeatureGroup, MapContainer, TileLayer } from "react-leaflet";
import { EditControl } from "react-leaflet-draw";
import { LatLngLiteral } from "leaflet";

import "leaflet/dist/leaflet.css";
import "leaflet-draw/dist/leaflet.draw.css";

const useStyles = createUseStyles({
  mapContainer: {
    height: '500px',
  },
});

export type PolygonMapLayer = {
  id: number;
  latlngs: LatLngLiteral[];
};

type Props = {
  setLayers: React.Dispatch<React.SetStateAction<PolygonMapLayer[]>>;
}

export const PolygonMap: React.FC<Props> = ({ setLayers }) => {
  const classes = useStyles();

  const _onCreate = (e: any) => {
    console.log(e);

    const { layerType, layer } = e;
    if (layerType === "polygon") {
      const { _leaflet_id } = layer;

      setLayers((layers) => [
        ...layers,
        { id: _leaflet_id, latlngs: layer.getLatLngs()[0] },
      ]);
    }
  };

  const _onEdited = (e: any) => {
    console.log(e);
    const {
      layers: { _layers },
    } = e;

    Object.values(_layers).forEach(({ _leaflet_id, editing }: any) => {
      setLayers((layers) =>
        layers.map((l) =>
          l.id === _leaflet_id
            ? { ...l, latlngs: editing.latlngs[0] }
            : l
        )
      );
    });
  };

  const _onDeleted = (e: any) => {
    console.log(e);
    const {
      layers: { _layers },
    } = e;

    Object.values(_layers).forEach(({ _leaflet_id }: any) => {
      setLayers((layers) => layers.filter((l) => l.id !== _leaflet_id));
    });
  };

  return (
    <MapContainer className={classes.mapContainer} center={[51.505, -0.09]} zoom={13} scrollWheelZoom={false}>
      <FeatureGroup>
        <EditControl
          position="topright"
          onCreated={_onCreate}
          onEdited={_onEdited}
          onDeleted={_onDeleted}
          draw={{
            rectangle: false,
            polyline: false,
            circle: false,
            circlemarker: false,
            marker: false,
          }}
        />
      </FeatureGroup>
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
    </MapContainer>
  )
};
