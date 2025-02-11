import React, { useEffect, useState } from "react"; 
import { createUseStyles } from 'react-jss';
import { FeatureGroup, MapContainer, TileLayer, useMapEvent } from "react-leaflet";
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

const defaultCenter: LatLngLiteral = { lat: 51.477, lng: 0 };
const defaultZoom = 4;

export const PolygonMap: React.FC<Props> = ({ setLayers }) => {
  const classes = useStyles();

  const [center, setCenter] = useState<LatLngLiteral>(() => {
    const savedCenter = localStorage.getItem("polygonMapCenter");
    return savedCenter ? JSON.parse(savedCenter) : defaultCenter;
  });

  const [zoom, setZoom] = useState<number>(() => {
    const savedZoom = localStorage.getItem("polygonMapZoom");
    return savedZoom ? JSON.parse(savedZoom) : defaultZoom;
  });

  useEffect(() => {
    localStorage.setItem("polygonMapCenter", JSON.stringify(center));
  }, [center]);

  useEffect(() => {
    localStorage.setItem("polygonMapZoom", JSON.stringify(zoom));
  }, [zoom]);

  const MyComponent = () => {
    const map = useMapEvent('moveend', () => {
      setCenter(map.getCenter());
      setZoom(map.getZoom());
    })
    return null
  }

  const _onCreate = (e: any) => {
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
    const {
      layers: { _layers },
    } = e;
    Object.values(_layers).forEach(({ _leaflet_id }: any) => {
      setLayers((layers) => layers.filter((l) => l.id !== _leaflet_id));
    });
  };

  return (
    <MapContainer
      className={classes.mapContainer}
      center={center}
      zoom={zoom}
      scrollWheelZoom={false}
    >
      <MyComponent />
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
