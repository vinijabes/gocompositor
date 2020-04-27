#include "gstreamer.h"
#include <gst/video/video.h>
#include <gst/audio/audio.h>
#include <gst/gstcaps.h>

void gstreamer_init() {
    gst_init(NULL, NULL);
}

GstElement* gstreamer_element_factory_make(const char* factoryname, const char* name){
    return GST_ELEMENT(gst_element_factory_make(factoryname, NULL));
}

int gstreamer_element_link(GstElement* src, GstElement* dest){
    return gst_element_link(src, dest);
}

void gstreamer_object_set(GstElement* object,const char* first_property_name, const char* arg){
    g_object_set(GST_OBJECT(object), first_property_name, arg, NULL);
}

typedef struct SignalHandlerUserData {
    int elementId;
    char* signal;
    GstElement *element;
} SignalHandlerUserData;

static void gstreamer_element_pad_added_signal_handler(GstElement *object, GstPad* new_pad, SignalHandlerUserData* data){
    goHandlePadAddedSignal(data->elementId, new_pad);
}

void gstreamer_element_pad_added_signal_connect(GstElement* element, int elementId){
    SignalHandlerUserData* s = calloc(1, sizeof(SignalHandlerUserData));
    s->element = element;
    s->elementId = elementId;

    s->signal = malloc(sizeof("pad-added"));
    strcpy(s->signal, "pad-added");

    g_signal_connect(element, "pad-added", G_CALLBACK(gstreamer_element_pad_added_signal_handler), s);
}

/* PIPELINE METHODS */

typedef struct BusMessageUserData {
    int pipelineId;
} BusMessageUserData;

GstPipeline* gstreamer_create_pipeline(const char* name){
    return (GstPipeline*) GST_BIN(gst_pipeline_new(name));
}

void gstreamer_pipeline_start(GstPipeline* pipeline){
    gst_element_set_state(GST_ELEMENT(pipeline), GST_STATE_PLAYING);
}

void gstreamer_pipeline_pause(GstPipeline *pipeline) {
    gst_element_set_state(GST_ELEMENT(pipeline), GST_STATE_PAUSED);
}

void gstreamer_pipeline_stop(GstPipeline *pipeline) {
    gst_element_set_state(GST_ELEMENT(pipeline), GST_STATE_NULL);
}

void gstreamer_pipeline_sendeos(GstPipeline *pipeline) {
    gst_element_send_event(GST_ELEMENT(pipeline), gst_event_new_eos());
}

void gstreamer_bin_add_element(GstPipeline *pipeline, GstElement* element){
    gst_bin_add(GST_BIN(pipeline), element);
}

gboolean gstreamer_bus_call(GstBus *bus, GstMessage *msg, gpointer user_data){
    BusMessageUserData *udata = (BusMessageUserData *)user_data;
    int pipelineId = udata->pipelineId;

    switch(GST_MESSAGE_TYPE(msg)){
        case GST_MESSAGE_EOS:
            goHandleBusMessage(msg, pipelineId);
            break;  
        case GST_MESSAGE_ERROR: {
            gchar *debug;
            GError *error;
            
            gst_message_parse_error(msg, &error, &debug);
            goPrint(debug);
            g_free(debug);
            g_error_free(error);
            
            goHandleBusMessage(msg, pipelineId);
            break;
        }
        case GST_MESSAGE_BUFFERING:
            goHandleBusMessage(msg,pipelineId);
            break;
        case GST_MESSAGE_STATE_CHANGED:
            goHandleBusMessage(msg, pipelineId);
            break;
        default:
            break;
    }

    return TRUE;
}

void gstreamer_pipeline_bus_watch(GstPipeline* pipeline, int pipelineId){
    BusMessageUserData *data = calloc(1, sizeof(BusMessageUserData));
    data->pipelineId = pipelineId;

    GstBus* bus = gst_pipeline_get_bus(pipeline);
    gst_bus_add_watch(bus, gstreamer_bus_call, data);
    gst_object_unref(bus);
}

/* PAD */

gboolean gstreamer_pad_link(GstPad* src, GstPad* dest){
    GstPadLinkReturn ret;
    ret = gst_pad_link(src, dest);
    
    return !GST_PAD_LINK_FAILED(ret);
}
