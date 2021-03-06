'use strict'

angular.module('sunglasses')
# renders a video element from vimeo or youtube
.directive('sunVideo', () ->
    getVideoWidget = (service, id) ->
        if Number(service) == 1
            return '<iframe 
            src="//www.youtube-nocookie.com/embed/'+id+'?rel=0" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>'
        else if Number(service) == 2
            return '<iframe src="//player.vimeo.com/video/'+id+'?color=2290d9" frameborder="0" 
            webkitallowfullscreen mozallowfullscreen allowfullscreen></iframe>'
    
    restrict: 'E',
    replace: true,
    template: '<div compile="widget"></div>',
    link: (scope, elem, attrs) ->
        scope.widget = getVideoWidget(attrs.service, attrs.videoId)
)